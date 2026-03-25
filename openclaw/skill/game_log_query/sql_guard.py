import re

FORBIDDEN = [
    "DELETE",
    "UPDATE",
    "INSERT",
    "DROP",
    "ALTER",
    "TRUNCATE"
]

def validate(sql):
    # 去掉前后空格
    sql_stripped = sql.strip()
    
    # 转大写方便匹配
    sql_upper = sql_stripped.upper()

    # 只允许 SELECT 开头
    if not sql_upper.startswith("SELECT"):
        raise Exception("只允许SELECT语句")

    # 禁止危险关键字
    for keyword in FORBIDDEN:
        if re.search(r"\b" + keyword + r"\b", sql_upper):
            raise Exception(f"危险SQL被阻止: {keyword}")

    # 必须查询日志表，严格匹配 item_log_YYYY-MM-DD
    if not re.search(r"\bITEM_LOG_\d{4}-\d{2}-\d{2}\b", sql_upper):
        raise Exception("只能查询日志表（item_log_YYYY-MM-DD）")

    return True
