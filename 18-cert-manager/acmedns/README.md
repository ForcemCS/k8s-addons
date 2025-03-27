## 安装

1. 安装系统依赖项

   ```
   sudo apt update
   sudo apt install python3 python3-venv libaugeas0
   ```

2. 设置 Python 虚拟环境

   ```
   sudo python3 -m venv /opt/certbot/
   sudo /opt/certbot/bin/pip install --upgrade pip
   ```

3. 安装 Certbot

   ```
   /opt/certbot/bin/pip install certbot
   pip install --no-cache-dir --index-url https://mirrors.aliyun.com/pypi/simple/ certbot  ##腾讯云
   ```

4. 安装acme-dns

   ```
   请[参考](https://github.com/joohoi/acme-dns-certbot-joohoi)
   ```

5. 使用acme-dns-certbot

   ```
   请[参考](https://www.digitalocean.com/community/tutorials/how-to-acquire-a-let-s-encrypt-certificate-using-dns-validation-with-acme-dns-certbot-on-ubuntu-18-04)
   
   ##如果有报错，设置#!/root/certbot/bin/python3
   ```

6. 命令

   ```
   certbot certonly --manual --manual-auth-hook /etc/letsencrypt/acme-dns-auth.py --preferred-challenges dns --debug-challenges --force-renewal -d \*.xxx.cn -d xxx.cn
   ```
