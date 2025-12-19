package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/xuri/excelize/v2"
)

// 配置信息
const (
	DuckDBFile      = "ro_prod.db"
	SecretDir       = `C:\Users\ForceCS\Desktop\go_project\go_basic\duckdb`
	MySQLHost       = "xxxxxxx"
	MySQLUser       = "readonly"
	MySQLPass       = "xxxxxxxxxx"
	MySQLPort       = xxxxxxx
	MySQLSecretName = "secret_ro"
	RemoteAlias     = "ro_prod_new"
)

func main() {
	// 默认日期逻辑：可传参或默认昨天
	targetDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if len(os.Args) > 1 {
		targetDate = os.Args[1]
	}

	log.Printf("🚀 任务开始 | 目标日期: %s", targetDate)

	if err := runWorkflow(targetDate); err != nil {
		log.Fatalf("❌ 任务失败: %v", err)
	}

	log.Println("✅ 任务全部完成！")
}

func runWorkflow(dateStr string) error {
	db, err := sql.Open("duckdb", DuckDBFile)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	reportTableName := "t_report_" + strings.ReplaceAll(dateStr, "-", "")

	// 1. 初始化环境与计算报表
	if err := calculateReport(ctx, db, dateStr, reportTableName); err != nil {
		return fmt.Errorf("计算失败: %w", err)
	}

	// 2. 导出到 Excel
	excelFileName := fmt.Sprintf("Daily_Report_%s.xlsx", strings.ReplaceAll(dateStr, "-", ""))
	if err := exportToExcel(db, reportTableName, excelFileName); err != nil {
		return fmt.Errorf("导出 Excel 失败: %w", err)
	}

	log.Printf("Excel 已生成: %s", excelFileName)
	return nil
}

func calculateReport(ctx context.Context, db *sql.DB, dateStr, tableName string) error {
	// --- 1. 安装并加载 MySQL 扩展 ---
	log.Println("🔧 安装并加载 mysql 扩展...")
	// 忽略安装错误（如果已安装会报错，不影响）
	db.ExecContext(ctx, "INSTALL mysql")
	if _, err := db.ExecContext(ctx, "LOAD mysql"); err != nil {
		return fmt.Errorf("加载 mysql 扩展失败: %w", err)
	}

	// --- 2. 设置环境 ---
	// 设置密钥目录
	db.ExecContext(ctx, fmt.Sprintf("SET secret_directory = '%s'", SecretDir))

	// 创建密钥
	createSecretSQL := fmt.Sprintf(`CREATE PERSISTENT SECRET IF NOT EXISTS %s (TYPE MYSQL, HOST '%s', USER '%s', PASSWORD '%s', PORT %d)`,
		MySQLSecretName, MySQLHost, MySQLUser, MySQLPass, MySQLPort)
	if _, err := db.ExecContext(ctx, createSecretSQL); err != nil {
		return fmt.Errorf("创建密钥失败: %w", err)
	}

	// 重新挂载数据库
	db.ExecContext(ctx, fmt.Sprintf("DETACH DATABASE IF EXISTS %s", RemoteAlias))
	attachSQL := fmt.Sprintf("ATTACH '' AS %s (TYPE mysql, SECRET %s)", RemoteAlias, MySQLSecretName)
	if _, err := db.ExecContext(ctx, attachSQL); err != nil {
		return fmt.Errorf("附加 MySQL 数据库失败: %w", err)
	}

	log.Println("✅ 数据库连接成功，开始执行计算逻辑...")

	// --- 3. 核心计算逻辑 (已修复语法错误) ---
	startTime := dateStr + " 00:00:00"
	endTime := dateStr + " 23:59:59"
	itemLogTable := fmt.Sprintf("%s.db_ro3_operation_log.\"item_log_%s\"", RemoteAlias, dateStr)

	reportSQL := fmt.Sprintf(`
		DROP TABLE IF EXISTS %s;
		CREATE TABLE %s AS
		WITH 
		params AS (
			SELECT TIMESTAMP '%s' as st, TIMESTAMP '%s' as et
		),
		active_uids AS (
			SELECT DISTINCT log.uid 
			FROM %s.db_ro3_operation_log.poli_island_log log
			CROSS JOIN params p
			WHERE log.time_stamp BETWEEN p.st AND p.et
		),
		stats_calc AS (
			SELECT 
				log.uid,
				MAX(CASE WHEN log.stagetype = 1 THEN log.stageid ELSE 0 END) AS max_stage_happy,
				MAX(CASE WHEN log.stagetype = 2 THEN log.stageid ELSE 0 END) AS max_stage_extreme,
				
				COUNT(*) FILTER (WHERE log.stagetype = 1 AND log.ismopup = 0) AS total_cnt_happy,
				COUNT(*) FILTER (WHERE log.stagetype = 1 AND log.ismopup = 0 AND log.time_stamp BETWEEN p.st AND p.et) AS today_cnt_happy,
				
				COUNT(*) FILTER (WHERE log.stagetype = 2 AND log.ismopup = 0) AS total_cnt_extreme,
				COUNT(*) FILTER (WHERE log.stagetype = 2 AND log.ismopup = 0 AND log.time_stamp BETWEEN p.st AND p.et) AS today_cnt_extreme,
				
				CAST(SUM(CASE WHEN log.ismopup = 1 THEN log.mopupcount ELSE 0 END) AS BIGINT) AS total_mopup,
				-- ▼▼▼ 之前报错的地方就是这里，现在已修复 ▼▼▼
				CAST(SUM(CASE WHEN log.ismopup = 1 AND log.time_stamp BETWEEN p.st AND p.et THEN log.mopupcount ELSE 0 END) AS BIGINT) AS today_mopup
			FROM %s.db_ro3_operation_log.poli_island_log log
			CROSS JOIN params p
			INNER JOIN active_uids a ON log.uid = a.uid
			GROUP BY log.uid
		),
		pay_stats AS (
			SELECT uid, SUM(amount) / 100.0 AS total_recharge
			FROM %s.db_ro3_sdk2.T_ORDER WHERE status = 2 GROUP BY uid
		),
		drop_stats AS (
			SELECT pl.uid, string_agg(il.itemid || ': ' || il.num, ', ') AS drop_info
			FROM %s.db_ro3_operation_log.poli_island_log pl
			CROSS JOIN params p
			INNER JOIN %s il ON pl.batchid = il.reason
			WHERE pl.ismopup = 1 AND pl.time_stamp BETWEEN p.st AND p.et
			GROUP BY pl.uid
		),
		happy_detail AS (
			SELECT uid, string_agg(stageid || ': ' || cnt, ', ' ORDER BY stageid ASC) AS stage_info
			FROM (
				SELECT uid, stageid, COUNT(*) as cnt 
				FROM %s.db_ro3_operation_log.poli_island_log log
				CROSS JOIN params p
				WHERE stagetype = 1 AND ismopup = 0 AND log.time_stamp BETWEEN p.st AND p.et
				GROUP BY uid, stageid
			) GROUP BY uid
		),
		extreme_detail AS (
			SELECT uid, string_agg(stageid || ': ' || cnt, ', ' ORDER BY stageid ASC) AS stage_info
			FROM (
				SELECT uid, stageid, COUNT(*) as cnt 
				FROM %s.db_ro3_operation_log.poli_island_log log
				CROSS JOIN params p
				WHERE stagetype = 2 AND ismopup = 0 AND log.time_stamp BETWEEN p.st AND p.et
				GROUP BY uid, stageid
			) GROUP BY uid
		)
		SELECT 
			t.uid, role.sid, role.nickname, role.viplv,
			COALESCE(pay.total_recharge, 0) AS total_recharge_amount,
			COALESCE(stats.max_stage_happy, 0) AS "欢乐冒险当前关卡id",
			COALESCE(stats.max_stage_extreme, 0) AS "极限挑战当前关卡id",
			COALESCE(stats.total_cnt_happy, 0) AS "欢乐冒险总挑战次数",
			COALESCE(stats.today_cnt_happy, 0) AS "欢乐冒险当日挑战总次数",
			COALESCE(stats.total_cnt_extreme, 0) AS "极限挑战总挑战次数",
			COALESCE(stats.today_cnt_extreme, 0) AS "极限挑战当日挑战总次数",
			COALESCE(stats.total_mopup, 0) AS "极限挑战历史总扫荡次数",
			COALESCE(stats.today_mopup, 0) AS "极限挑战当日总扫荡次数",
			COALESCE(d.drop_info, '') AS "当日扫荡掉落物品ID:数量",
			COALESCE(ed.stage_info, '') AS "关卡极限挑战挑战次数",
			COALESCE(hd.stage_info, '') AS "关卡欢乐冒险挑战次数",
			COALESCE(role.power, 0) AS power
		FROM active_uids t
		LEFT JOIN %s.db_ro3_operation_log.SNAP_ROLE role ON t.uid = role.uid
		LEFT JOIN stats_calc stats ON t.uid = stats.uid
		LEFT JOIN pay_stats pay ON t.uid = pay.uid
		LEFT JOIN drop_stats d ON t.uid = d.uid
		LEFT JOIN happy_detail hd ON t.uid = hd.uid
		LEFT JOIN extreme_detail ed ON t.uid = ed.uid;
	`, tableName, tableName, startTime, endTime,
		RemoteAlias,               // active_uids
		RemoteAlias,               // stats_calc
		RemoteAlias,               // pay_stats
		RemoteAlias, itemLogTable, // drop_stats
		RemoteAlias, // happy_detail
		RemoteAlias, // extreme_detail
		RemoteAlias) // final select

	if _, err := db.ExecContext(ctx, reportSQL); err != nil {
		return fmt.Errorf("执行报表 SQL 失败: %w", err)
	}
	return nil
}

func exportToExcel(db *sql.DB, tableName, fileName string) error {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet) // 默认创建的 Sheet1

	// 1. 写入表头
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}

	// 2. 写入数据
	rowIdx := 2
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("⚠️ 扫描行失败: %v", err)
			continue
		}

		for i, val := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
			// 处理数据类型转换
			f.SetCellValue(sheet, cell, val)
		}
		rowIdx++
	}

	// 3. 简单的样式美化（加粗表头）
	style, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	f.SetRowStyle(sheet, 1, 1, style)

	return f.SaveAs(fileName)
}
