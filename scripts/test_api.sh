#!/bin/bash

echo "🚀 向预言机发送跨链验证请求..."
echo "====================================="

# 默认测试一个正确的哈希，如果运行脚本时传了参数，就用参数里的哈希
HASH=${1:-"Qm1234567890abcdef"}

echo "🔍 目标哈希: $HASH"
curl -s "http://localhost:8080/api/verify?hash=$HASH" | jq . || curl -s "http://localhost:8080/api/verify?hash=$HASH"
echo -e "\n====================================="
