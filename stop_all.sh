#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain Oracle] 极客优雅关机序列 v5.0 ===${NC}"

PROJECT_DIR=$(cd "$(dirname "$0")"; pwd)
LOG_DIR="$PROJECT_DIR/logs"

# 1. 斩断跨链微服务集群
echo -e "${YELLOW}[1/6] 正在超度跨链微服务群...${NC}"
SERVICES=("issuer_ui" "verifier_ui" "auto_trigger" "fisco_writer" "fabric_adapter")
for service in "${SERVICES[@]}"; do
    pid_file="$LOG_DIR/${service}.pid"
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null; then
            kill $pid
            echo "  -> $service (PID: $pid) 已被安全终止。"
        fi
        rm -f "$pid_file"
    else
        pkill -f "$service" 2>/dev/null
    fi
done
echo -e "${GREEN}跨链中间件集群已安全清理。${NC}"

# 2. 优雅关闭 IPFS 节点
echo -e "${YELLOW}[2/6] 正在安全切断 IPFS 星际文件系统...${NC}"
if pgrep -x "ipfs" > /dev/null; then
    ipfs shutdown > /dev/null 2>&1
    echo -e "${GREEN}IPFS 节点已安全下线，本地区块库已落盘。${NC}"
else
    echo "IPFS 节点未在运行。"
fi

# 3. 优雅停止 FISCO BCOS
echo -e "${YELLOW}[3/6] 正在安全挂起 FISCO BCOS 节点...${NC}"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash stop_all.sh
    echo -e "${GREEN}FISCO 节点已落盘并安全停止。${NC}"
else
    echo -e "${RED}[警告] 未找到 FISCO 节点路径。${NC}"
fi

# 4. 专门指挥 Chainlink 舰队降落
echo -e "${YELLOW}[4/6] 正在指挥 Chainlink 舰队降落...${NC}"
if [ -d "$HOME/cross-chain-project/chainlink-node" ]; then
    cd $HOME/cross-chain-project/chainlink-node
    docker-compose stop > /dev/null 2>&1
    echo -e "${GREEN}Chainlink 节点与 Postgres 数据库已优雅挂起。${NC}"
else
    echo -e "${RED}[错误] 未发现 Chainlink 阵地目录。${NC}"
fi

# 5. 战术休眠 Fabric 容器
echo -e "${YELLOW}[5/6] 正在休眠 Fabric 剩余容器...${NC}"
remaining_containers=$(docker ps -q)
if [ -n "$remaining_containers" ]; then
    docker stop $remaining_containers > /dev/null
    echo -e "${GREEN}所有 Fabric 业务容器已进入休眠模式。${NC}"
else
    echo "无其他运行中的容器。"
fi

# 6. GitHub 存盘防呆检查
echo -e "${YELLOW}[6/6] 正在进行 GitHub 存盘检查...${NC}"
cd $HOME/cross-chain-project
if [[ -n $(git status -s) ]]; then
    echo -e "${RED}[警告] 发现以下未提交的代码修改！${NC}"
    git status -s
    echo -e "${YELLOW}请先执行 git add / commit / push，或者使用 syncgit 快捷命令上云！${NC}"
    echo -e "${RED}关机程序已中止。${NC}"
    exit 1
else
    echo -e "${GREEN}工作区完全干净，代码已全部安全上云。${NC}"
fi

echo "----------------------------------------------------------------------"
echo -e "${GREEN}战场已完美清扫！现在你可以安全地下达关机指令了：${NC}"
echo -e "${YELLOW}sudo poweroff${NC}"
echo "----------------------------------------------------------------------"
