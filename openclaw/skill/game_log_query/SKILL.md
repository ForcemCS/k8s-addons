---
name: game_log_query
description: 查询游戏玩家道具、奖励、消耗、充值等日志。适合客服自然语言排查玩家行为。
tools:
  - python
---

## 数据库信息

- 数据库：`db_ro3_operation_log`
- 表结构：按天分表，表名格式 `item_log_YYYY-MM-DD`
- 示例：`item_log_2026-03-08`

### 字段说明

| 字段        | 类型     | 含义                 |
| ----------- | -------- | -------------------- |
| id          | int      | 主键                 |
| uid         | varchar  | 玩家ID               |
| sid         | int      | 区服ID               |
| where       | varchar  | 消耗渠道 / 来源      |
| itemid      | int      | 道具ID               |
| num         | int      | 数量                 |
| left        | int      | 剩余数量             |
| change_type | int      | 类型：1=消耗，2=获得 |
| reason      | varchar  | 原因                 |
| time_stamp  | datetime | 时间戳               |

---

## 查询规则

1. **奖励获取** → `change_type=2`  
2. **道具消耗** → `change_type=1`  
3. **充值** → `change_type=充值表字段`（可扩展）  
4. **邮件领取** → `where='mail'`  
5. **时间模糊匹配**  
   - “早上10点左右” → ±30分钟  
   - “下午3点半” → ±30分钟  
6. **分表选择**  
   - 根据日期自动选 `item_log_YYYY-MM-DD`  
   - 如果表不存在，提示“日志表不存在”

---

## 安全限制

- 只允许 `SELECT`  
- 只允许查询 `item_log_*` 表  
- 禁止 DROP / DELETE / UPDATE / INSERT / ALTER / TRUNCATE  
- 查询最大时间范围限制：2小时  

---

## 工具调用

虚拟环境在venv中 ,生成 SQL 后调用：

```bash
python sql_executor.py "<SQL>"
```

- `sql_executor.py` 内置 SQL 安全过滤  
- 返回字典形式查询结果  
- 可由 LLM 总结成客服可读文本  

---

## 示例

### 1. 查询奖励

**客服输入：**  
玩家1000331710004 8号10点左右获得奖励

**生成 SQL：**

```sql
SELECT *
FROM item_log_2026-03-08
WHERE uid='1000331710004'
AND change_type=2
AND time_stamp BETWEEN '2026-03-08 09:30:00' AND '2026-03-08 10:30:00'
ORDER BY time_stamp;
```

### 2. 查询消耗

**客服输入：**  
玩家1000331710004 8号10点左右的消耗

**生成 SQL：**

```sql
SELECT *
FROM item_log_2026-03-08
WHERE uid='1000331710004'
AND change_type=1
AND time_stamp BETWEEN '2026-03-08 09:30:00' AND '2026-03-08 10:30:00'
ORDER BY time_stamp;
```

### 3. 查询邮件领取

**客服输入：**  
玩家1000331710004 8号10点左右领取了哪些邮件

**生成 SQL：**

```sql
SELECT *
FROM item_log_2026-03-08
WHERE uid='1000331710004'
AND where='mail'
AND time_stamp BETWEEN '2026-03-08 09:30:00' AND '2026-03-08 10:30:00'
ORDER BY time_stamp;
```

---

## 数据导出

如果用户要求 **导出查询结果 / 导出数据 / 下载数据 / 保存查询结果**：

1. 将查询结果导出为 CSV 文件
2. 调用 `s3_upload` skill 上传文件
3. 告诉用户下载地址

RustFS Web UI：http://10.10.0.202:9101

Bucket：data-report

示例流程：

1. 生成文件  xxx.csv
2. 调用上传 skill
   + s3_upload xxx.csv
3.  回复用户
   + 数据已经导出并上传。请在 RustFS UI 查看：http://10.10.0.202:9101

## 扩展说明

- 任何自然语言中涉及时间、玩家ID、行为类型（奖励、消耗、充值、邮件领取），LLM 都可解析并生成 SQL  
- 执行后可由 LLM 自动总结为客服可读文本  
