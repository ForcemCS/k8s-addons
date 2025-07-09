### 服务端配置

```
sudo apt update
sudo apt install wireguard
```

```
umask 077 # 确保文件权限为 077
wg genkey | tee privatekey_server | wg pubkey > publickey_server
# privatekey_server 保存服务器私钥 (重要，不能泄露)
# publickey_server 保存服务器公钥 (需要给客户端)
```

```
root@debian:~# cat   /etc/wireguard/wg0.conf
[Interface]
PrivateKey = xxxxx1
Address = 192.168.0.1/24 # WireGuard 服务器在隧道内的 IP，通常是隧道的第一个IP
ListenPort = 51820
#PostUp = sysctl -w net.ipv4.ip_forward=1; iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o enp5s0 -j MASQUERADE
#PostDown = sysctl -w net.ipv4.ip_forward=0; iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o enp5s0 -j MASQUERADE
PostUp = sysctl -w net.ipv4.ip_forward=1; iptables -I FORWARD 1 -i %i -j ACCEPT; iptables -I FORWARD 1 -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o enp5s0 -j MASQUERADE
PostDown = sysctl -w net.ipv4.ip_forward=0; iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o enp5s0 -j MASQUERADE

[Peer]
PublicKey = homexx
AllowedIPs = 192.168.0.2/32 # 您的家用电脑在隧道内的 IP
PersistentKeepalive = 25
#iptables -I INPUT 1 -i wg0 -j ACCEPT

[Peer]
PublicKey = xxxxxx3
AllowedIPs = 192.168.0.3/32 # 您的家用电脑在隧道内的 IP
PersistentKeepalive = 25

[Peer]
PublicKey = xxxxx4
AllowedIPs = 192.168.0.4/32 # 您的家用电脑在隧道内的 IP
PersistentKeepalive = 25

```

```
net.ipv4.ip_forward=1
```

```
sudo systemctl start wg-quick@wg0
sudo wg show # 检查 WireGuard 状态
```

### 客户端配置

```
[Interface]
PrivateKey = home
Address = 192.168.0.4/32 

[Peer]
PublicKey = serrverxxxx
Endpoint = 118.242.18.133:51820
AllowedIPs = 192.168.0.0/24,10.10.0.0/16
PersistentKeepalive = 25 # 建议开启，用于维持 NAT 映射，避免连接被家庭路由器超时断开
```

