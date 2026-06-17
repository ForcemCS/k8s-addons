package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Fluent Bit http 输出(Format json)发来的记录结构,只取我们关心的字段
type record struct {
	Log        string `json:"log"`
	Kubernetes struct {
		NamespaceName string            `json:"namespace_name"`
		PodName       string            `json:"pod_name"`
		Labels        map[string]string `json:"labels"`
	} `json:"kubernetes"`
}

var (
	webhookURL = os.Getenv("WECHAT_WEBHOOK_URL")
	cooldown   = getDur("COOLDOWN_SECONDS", 300)  // 同一服务的告警冷却,默认 5 分钟
	maxBytes   = getInt("MAX_CONTENT_BYTES", 1500) // 企业微信 text 上限 2048 字节,留余量给头部

	// instance 白名单,存当前生效的 map[string]bool;热加载时原子替换
	instances atomic.Value

	mu       sync.Mutex
	lastSent = map[string]time.Time{}
	suppress = map[string]int{}
)

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getDur(k string, defSec int) time.Duration {
	return time.Duration(getInt(k, defSec)) * time.Second
}

// 同时兼容逗号和换行分隔,方便 ConfigMap 里一行一个写
func parseSet(s string) map[string]bool {
	m := map[string]bool{}
	s = strings.ReplaceAll(s, "\n", ",")
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			m[p] = true
		}
	}
	return m
}

// 读取白名单原始内容:优先文件(ALERT_INSTANCES_FILE),否则用环境变量 ALERT_INSTANCES
func readRaw() (string, bool) {
	if f := os.Getenv("ALERT_INSTANCES_FILE"); f != "" {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Println("读取白名单文件失败,保持现状:", err)
			return "", false
		}
		return string(b), true
	}
	return os.Getenv("ALERT_INSTANCES"), true
}

// 归一化成有序、去重的字符串,用于判断内容是否真的变了
func normalize(raw string) string {
	set := parseSet(raw)
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func logInstances(raw string) {
	if len(parseSet(raw)) == 0 {
		log.Println("白名单为空 -> 所有 instance 都告警")
	} else {
		log.Println("当前告警白名单:", normalize(raw))
	}
}

// 后台定期重读白名单,变化时原子替换并打印
func watchInstances(initialRaw string) {
	last := normalize(initialRaw)
	t := time.NewTicker(getDur("RELOAD_SECONDS", 15))
	defer t.Stop()
	for range t.C {
		raw, ok := readRaw()
		if !ok || normalize(raw) == last {
			continue
		}
		last = normalize(raw)
		instances.Store(parseSet(raw))
		log.Println("白名单已热更新")
		logInstances(raw)
	}
}

func main() {
	if webhookURL == "" {
		log.Fatal("环境变量 WECHAT_WEBHOOK_URL 未设置")
	}
	raw, _ := readRaw()
	instances.Store(parseSet(raw))
	logInstances(raw)
	go watchInstances(raw)

	http.HandleFunc("/skynet-alert", handle)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	log.Println("skynet-alert-relay listening on :8080, cooldown =", cooldown)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var recs []record
	if err := json.Unmarshal(body, &recs); err != nil {
		var one record
		if json.Unmarshal(body, &one) == nil {
			recs = []record{one}
		}
	}

	allow, _ := instances.Load().(map[string]bool)

	for _, rec := range recs {
		ns := rec.Kubernetes.NamespaceName
		app := rec.Kubernetes.Labels["app.kubernetes.io/name"]
		inst := rec.Kubernetes.Labels["app.kubernetes.io/instance"]

		// 白名单过滤:设置了名单且当前 instance 不在其中 -> 跳过,不告警
		if len(allow) > 0 && !allow[inst] {
			continue
		}

		key := ns + "/" + app + "/" + inst

		mu.Lock()
		lastT, seen := lastSent[key]
		if seen && time.Since(lastT) < cooldown {
			suppress[key]++ // 冷却期内只计数,不发
			mu.Unlock()
			continue
		}
		n := suppress[key]
		suppress[key] = 0
		lastSent[key] = time.Now()
		mu.Unlock()

		go send(ns, app, inst, rec.Kubernetes.PodName, rec.Log, n)
	}
	w.WriteHeader(http.StatusOK)
}

func send(ns, app, inst, pod, logText string, suppressed int) {
	header := "🚨 游戏服异常告警(stack traceback)\n" +
		"服务: " + app + " / " + inst + "  (ns=" + ns + ")\n" +
		"Pod: " + pod + "\n"
	if suppressed > 0 {
		header += "⚠️ 距上次告警期间该服务还累计出现 " + strconv.Itoa(suppressed) + " 次\n"
	}
	header += "----------------\n"

	content := truncate(header+logText, maxBytes)

	payload := map[string]any{
		"msgtype": "text",
		"text":    map[string]any{"content": content},
	}
	b, _ := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println("发送企业微信失败:", err)
		return
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	log.Printf("已发送告警 service=%s/%s resp=%s", app, inst, string(rb))
}

// 按字节上限截断,且不切断多字节(中文)字符
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	var b []byte
	for _, c := range s {
		cb := []byte(string(c))
		if len(b)+len(cb) > max {
			break
		}
		b = append(b, cb...)
	}
	return string(b) + "\n...(内容过长已截断,完整日志见 Loki)"
}
