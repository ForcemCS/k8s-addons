import sys
import re
import mysql.connector
from mysql.connector import errorcode
from sql_guard import validate

def fix_table_quotes(query):
    """
    自动修复 SQL 中未加反引号的 item_log_YYYY-MM-DD 表名。
    例如将 item_log_2026-03-08 替换为 `item_log_2026-03-08`
    """
    # 匹配 item_log_ 后面跟着日期格式且未加反引号的情况
    pattern = r'(?<!`)(item_log_\d{4}-\d{2}-\d{2})(?!`)'
    return re.sub(pattern, r'`\1`', query)

def execute_query():
    if len(sys.argv) < 2:
        print("错误: 未提供 SQL 查询语句")
        sys.exit(1)

    # 1. 获取并预处理 SQL
    raw_sql = sys.argv[1]
    
    # 修复可能导致语法错误的表名
    processed_sql = fix_table_quotes(raw_sql)

    # 2. 安全校验 (调用你现有的 sql_guard)
    try:
        validate(processed_sql)
    except Exception as e:
        print(f"安全校验未通过: {e}")
        sys.exit(1)

    # 3. 数据库连接配置
    config = {
        'host': "xxx",
        'user': "root",
        'password': "xxx",
        'port': xxx,
        'database': "xxxx",
        'charset': 'utf8mb4',
        'collation': 'utf8mb4_general_ci',
        'use_pure': True  # 提高对复杂环境的兼容性
    }

    conn = None
    try:
        conn = mysql.connector.connect(**config)
        cursor = conn.cursor(dictionary=True)

        # 4. 执行查询
        cursor.execute(processed_sql)
        rows = cursor.fetchall()

        # 5. 输出结果
        if not rows:
            print("没有查询到记录")
        else:
            # 以标准格式输出，方便 OpenClaw 解析
            for r in rows:
                print(r)

    except mysql.connector.Error as err:
        if err.errno == errorcode.ER_ACCESS_DENIED_ERROR:
            print("错误: 数据库用户名或密码错误")
        elif err.errno == errorcode.ER_BAD_DB_ERROR:
            print("错误: 数据库不存在")
        elif err.errno == 1064:
            print(f"SQL 语法错误 [1064]: {err.msg}\n生成的 SQL: {processed_sql}")
        elif err.errno == 1146:
            print(f"错误: 指定的日志表不存在 (当天可能无日志)")
        else:
            print(f"数据库错误: {err}")
    except Exception as e:
        print(f"系统错误: {e}")
    finally:
        if conn and conn.is_connected():
            cursor.close()
            conn.close()

if __name__ == "__main__":
    execute_query()
