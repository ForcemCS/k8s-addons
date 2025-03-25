#!/bin/bash

# 确保传入的参数
if [ $# -lt 2 ]; then
  echo -e "\e[31mUsage: $0 <工作目录> <证书文件的名称>\e[0m"
  exit 1
fi

# 获取传入的参数
WORKDIR=$1
CLIENT_NAME=$2

# 确保工作目录存在
mkdir -p ${WORKDIR}

# 步骤 1：生成客户端私钥
echo -e "\e[31m步骤 1：生成客户端私钥\e[0m"
openssl genrsa -out ${WORKDIR}/${CLIENT_NAME}.key 2048

# 步骤 2：创建客户端证书签名请求 (CSR) 配置文件
echo -e "\e[31m步骤 2：创建客户端证书签名请求 (CSR) 配置文件\e[0m"
cat > ${WORKDIR}/${CLIENT_NAME}-csr.conf <<EOF
[req]
default_bits = 2048
prompt = no
encrypt_key = yes
default_md = sha256
distinguished_name = client_auth
req_extensions = v3_req

[ client_auth ]
O = system:nodes
CN = system:node:${CLIENT_NAME}-svc

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth                           # 关键：这里只指定 clientAuth
# 不需要 subjectAltName
EOF

# 步骤 3：生成客户端 CSR
echo -e "\e[31m步骤 3：生成客户端 CSR\e[0m"
openssl req -new -key ${WORKDIR}/${CLIENT_NAME}.key -out ${WORKDIR}/${CLIENT_NAME}.csr -config ${WORKDIR}/${CLIENT_NAME}-csr.conf

# 步骤 4：创建客户端 CSR YAML 文件
echo -e "\e[31m步骤 4：创建客户端 CSR YAML 文件\e[0m"
cat > ${WORKDIR}/${CLIENT_NAME}-csr.yaml <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: ${CLIENT_NAME}.csr  # CSR 资源的名称，可以自定义
spec:
  signerName: kubernetes.io/kube-apiserver-client    # 关键：签名者名称
  expirationSeconds: 157680000  
  request: $(cat ${WORKDIR}/${CLIENT_NAME}.csr | base64 | tr -d '\n')
  usages:
  - client auth                                      # 关键：这里指定 client auth
EOF

# 步骤 5：向 Kubernetes 发送 CSR
echo -e "\e[31m步骤 5：向 Kubernetes 发送 CSR\e[0m"
kubectl create -f ${WORKDIR}/${CLIENT_NAME}-csr.yaml

# 步骤 6：批准 CSR
echo -e "\e[31m步骤 6：批准 CSR\e[0m"
kubectl certificate approve ${CLIENT_NAME}.csr

echo -e "\e[32m客户端证书请求已成功发送并批准！\e[0m"

# 步骤 7：检查 CSR 状态并验证证书签发结果
echo -e "\e[31m步骤 7：检查 CSR 状态并验证证书签发结果\e[0m"
CSR_STATUS=$(kubectl get csr ${CLIENT_NAME}.csr -o jsonpath='{.status.conditions[*].type}')

# 判断是否有 Failed 状态
if [[ "$CSR_STATUS" =~ "Failed" ]]; then
  echo -e "\e[31m证书签发失败！\e[0m"
  exit 1
fi

# 如果没有 Failed，表示签发成功
echo -e "\e[32m证书签发成功！\e[0m"
