import argparse
import pymysql
import json
import datetime
from decimal import Decimal

# 处理日期和高精度数字的JSON序列化
class CustomEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (datetime.datetime, datetime.date)):
            return obj.strftime('%Y-%m-%d %H:%M:%S')
        if isinstance(obj, Decimal):
            return float(obj)
        return super().default(obj)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--sql', type=str, required=True, help="要执行的完整 SQL 语句")
    args = parser.parse_args()

    conn = pymysql.connect(
        host='a', 
        port=b,       
        user='root', 
        password='c', 
        database='d', 
        cursorclass=pymysql.cursors.DictCursor
    )

    try:
        with conn.cursor() as cursor:
            # 执行大模型传过来的 SQL
            cursor.execute(args.sql)
            result = cursor.fetchall()
            # 将查询结果返回给大模型
            print(json.dumps(result, ensure_ascii=False, cls=CustomEncoder))
    except Exception as e:
        print(f"SQL执行错误: {e}")
    finally:
        conn.close()

if __name__ == '__main__':
    main()
