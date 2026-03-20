#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain] 端到端自动化开火序列 v3.0 ===${NC}"

# 1. 向控制台发射查证请求并强制退出
echo -e "${YELLOW}[1/3] 正在向 FISCO BCOS 控制台 (Group 1) 发射跨链查证子弹...${NC}"
cd $HOME/console

CMD='call CrossChainClient 0x77580befe1c74597dfdd626d6cced71807b868d4 requestVerification "QmTestHash123456789"'

# 核心修复：使用 printf 传入指令，并紧跟一个 quit 强制结束控制台交互进程
OUTPUT=$(printf '%s\nquit\n' "$CMD" | bash start.sh 1 2>&1)

# 2. 截获 Transaction Hash
echo -e "${YELLOW}[2/3] 正在雷达日志中扫描 Transaction Hash...${NC}"
# 使用正则表达式精准抓取 64 位哈希字符串
TX_HASH=$(echo "$OUTPUT" | grep -o '0x[0-9a-fA-F]\{64\}' | head -n 1)

if [ -z "$TX_HASH" ]; then
    echo -e "${RED}[错误] 未能从控制台截获有效哈希，子弹可能卡壳。输出日志如下：${NC}"
    echo "$OUTPUT"
    exit 1
fi

echo -e "${GREEN}成功截获 Transaction Hash: $TX_HASH${NC}"

# 3. 预言机自动穿刺与回写
echo -e "${YELLOW}[3/3] 将哈希装填至预言机中枢，开始跨链截获与 Fabric 回写...${NC}"
cd $HOME/cross-chain-project/listener
go run main.go "$TX_HASH"

echo "----------------------------------------------------------------------"
echo -e "${GREEN}全自动化跨链查证流程执行完毕！${NC}"
echo "----------------------------------------------------------------------"
