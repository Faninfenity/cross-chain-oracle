#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain Oracle] 极客优雅关机序列 v2.1 ===${NC}"

# 1. 斩断预言机微服务 (原生 Go 进程)
echo -e "${YELLOW}[1/5] 正在超度预言机后台进程...${NC}"
pkill -f "adapter.go" 2>/dev/null
pkill -f "main.go" 2>/dev/null
echo -e "${GREEN}预言机中间件已安全清理。${NC}"

# 2. 优雅停止 FISCO BCOS
echo -e "${YELLOW}[2/5] 正在安全挂起 FISCO BCOS 节点...${NC}"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash stop_all.sh
    echo -e "${GREEN}FISCO 节点已落盘并安全停止。${NC}"
else
    echo -e "${RED}[警告] 未找到 FISCO 节点路径。${NC}"
fi

# 3. 专门指挥 Chainlink 舰队降落
echo -e "${YELLOW}[3/5] 正在指挥 Chainlink 舰队降落...${NC}"
if [ -d "$HOME/cross-chain-project/chainlink-node" ]; then
    cd $HOME/cross-chain-project/chainlink-node
    # 使用 stop 而非 down，保留容器状态，下次启动快
    docker-compose stop > /dev/null 2>&1
    echo -e "${GREEN}Chainlink 节点与 Postgres 数据库已优雅挂起。${NC}"
else
    echo -e "${RED}[错误] 未发现 Chainlink 阵地目录。${NC}"
fi

# 4. 战术休眠 Fabric 容器
echo -e "${YELLOW}[4/5] 正在休眠 Fabric 剩余容器...${NC}"
remaining_containers=$(docker ps -q)
if [ -n "$remaining_containers" ]; then
    docker stop $remaining_containers > /dev/null
    echo -e "${GREEN}所有 Fabric 业务容器已进入休眠模式。${NC}"
else
    echo "无其他运行中的容器。"
fi

# 5. GitHub 存盘防呆检查
echo -e "${YELLOW}[5/5] 正在进行 GitHub 存盘检查...${NC}"
cd $HOME/cross-chain-project
if [[ -n $(git status -s) ]]; then
    echo -e "${RED}[警告] 发现以下未提交的代码修改！${NC}"
    git status -s
    echo -e "${YELLOW}请先执行 git add / commit / push，确保心血上云后再关机！${NC}"
    echo -e "${RED}关机程序已中止。${NC}"
    exit 1
else
    echo -e "${GREEN}工作区完全干净，代码已全部安全上云。${NC}"
fi

echo "----------------------------------------------------------------------"
echo -e "${GREEN}战场已完美清扫！现在你可以安全地下达关机指令了：${NC}"
echo -e "${YELLOW}sudo poweroff${NC}"
echo "----------------------------------------------------------------------"
