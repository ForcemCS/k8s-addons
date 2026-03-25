import argparse
import smtplib
import json
import os
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email.header import Header
from email import encoders

# ================= 配置区 =================
SENDER_EMAIL = 'xxx@example.org'
SENDER_PASSWORD = 'xxxxxxxxx' 
SMTP_SERVER = 'smtp.exmail.qq.com'
SMTP_PORT = 465
# ==========================================

def send_email(to_emails_str, subject, body, attachment_path=None):
    """执行实际的邮件发送动作，支持多收件人和附件"""
    
    to_list = [email.strip() for email in to_emails_str.split(',') if email.strip()]
    if not to_list:
        return {"status": "error", "message": "未提供有效的收件人邮箱地址。"}

    body_html = body.replace('\\n', '<br>').replace('\n', '<br>')
    
    html_template = f"""
    <html>
    <head>
    <style>
        body {{ font-family: 'Microsoft YaHei', Arial, sans-serif; line-height: 1.6; color: #333333; background-color: #f4f5f7; padding: 20px; }}
        .email-container {{ max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e5e5e5; }}
        .email-header {{ background-color: #2b7cff; color: #ffffff; padding: 20px; font-size: 18px; font-weight: bold; border-bottom: 2px solid #1a60d6; }}
        .email-body {{ padding: 30px 20px; font-size: 15px; color: #444444; }}
        .email-footer {{ background-color: #fafafa; padding: 15px 20px; font-size: 12px; color: #888888; text-align: right; border-top: 1px solid #eeeeee; }}
    </style>
    </head>
    <body>
        <div class="email-container">
            <div class="email-header">
                📝 {subject}
            </div>
            <div class="email-body">
                {body_html}
            </div>
            <div class="email-footer">
                此邮件由 <b>未来互娱 AI 智能办公助手</b> 自动发送
            </div>
        </div>
    </body>
    </html>
    """

    # 构造邮件对象 (MIMEMultipart 默认就是 mixed，支持正文+附件)
    msg = MIMEMultipart()
    msg['From'] = Header(f"智能办公助手 <{SENDER_EMAIL}>", 'utf-8')
    msg['To'] = ", ".join(to_list)
    msg['Subject'] = Header(subject, 'utf-8')
    
    # 1. 挂载 HTML 正文
    msg.attach(MIMEText(html_template, 'html', 'utf-8'))

    # 2. 【核心新增】：挂载附件
    if attachment_path:
        # 检查文件是否存在
        if not os.path.exists(attachment_path):
            return {"status": "error", "message": f"发送失败：找不到附件文件 {attachment_path}"}
        
        try:
            # 以二进制读取文件
            with open(attachment_path, 'rb') as f:
                part = MIMEBase('application', 'octet-stream')
                part.set_payload(f.read())
            
            # Base64 编码
            encoders.encode_base64(part)
            
            # 获取文件名并处理中文编码问题
            filename = os.path.basename(attachment_path)
            part.add_header('Content-Disposition', 'attachment', filename=Header(filename, 'utf-8').encode())
            
            # 将附件添加到邮件对象中
            msg.attach(part)
        except Exception as e:
            return {"status": "error", "message": f"附件处理失败: {str(e)}"}

    try:
        server = smtplib.SMTP_SSL(SMTP_SERVER, SMTP_PORT)
        server.login(SENDER_EMAIL, SENDER_PASSWORD)
        server.sendmail(SENDER_EMAIL, to_list, msg.as_string())
        server.quit()
        
        att_msg = f" (包含附件: {os.path.basename(attachment_path)})" if attachment_path else ""
        return {"status": "success", "message": f"邮件已成功发送给 {len(to_list)} 个人{att_msg}。"}
        
    except Exception as e:
        return {"status": "error", "message": f"发送失败: {str(e)}"}

def main():
    parser = argparse.ArgumentParser(description="企业微信邮件发送执行脚本")
    parser.add_argument('--to', type=str, required=True, help="收件人精准邮箱地址，多个用英文逗号分隔")
    parser.add_argument('--subject', type=str, required=True, help="邮件主题")
    parser.add_argument('--body', type=str, required=True, help="邮件正文")
    # 新增 --attachment 参数，非必填 (required=False)
    parser.add_argument('--attachment', type=str, required=False, default=None, help="附件的绝对路径")
    
    args = parser.parse_args()

    # 传入附件路径
    result = send_email(args.to, args.subject, args.body, args.attachment)
    print(json.dumps(result, ensure_ascii=False))

if __name__ == '__main__':
    main()
