#!/bin/bash
set -e

# --- 1. 配置变量 (保持不变) ---
REGISTRY="example.com"
NAMESPACE="micros"
IMAGE_NAME="ro3-api2"
IMAGE_TAG=$(git rev-parse --short HEAD)
FULL_IMAGE_REF="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}"

# --- 2. 定义远程缓存的位置 ---
# 使用一个特定的标签来存储远程构建缓存
CACHE_REF="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:buildcache"

# --- 3. 执行现代的 buildx 构建命令 ---
echo "=========================================================="
echo "使用 Docker Buildx 开始构建:"
echo "  ${FULL_IMAGE_REF}"
echo "远程仓库缓存: ${CACHE_REF}"
echo "=========================================================="

docker buildx build --cache-to "type=registry,ref=${CACHE_REF}" --cache-from "type=registry,ref=${CACHE_REF}"  --push --tag "${FULL_IMAGE_REF}"  .

echo "=========================================================="
echo "构建和推送成功！"
echo "镜像: ${FULL_IMAGE_REF}"
echo "=========================================================="
