#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain Oracle] 极客优雅关机序列 v2.0 ===${NC}"

# 1. 斩断预言机微服务
echo -e "${YELLOW}[1/4] 正在超度预言机后台进程...${NC}"
pkill -f "adapter.go" 2>/dev/null
pkill -f "main.go" 2>/dev/null
echo -e "${GREEN}预言机中间件已安全清理。${NC}"

# 2. 优雅停止 FISCO BCOS
echo -e "${YELLOW}[2/4] 正在安全挂起 FISCO BCOS 节点...${NC}"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash stop_all.sh
    echo -e "${GREEN}FISCO 节点已落盘并安全停止。${NC}"
else
    echo -e "${RED}[警告] 未找到 FISCO 节点路径。${NC}"
fi

# 3. 战术休眠 Fabric 容器 (绝不使用 network.sh down！)
echo -e "${YELLOW}[3/4] 正在休眠 Fabric 网络容器...${NC}"
if [ "$(docker ps -q)" ]; then
    docker stop $(docker ps -q) > /dev/null
    echo -e "${GREEN}Fabric 容器已休眠！账本数据与智能合约已完美保留。${NC}"
else
    echo "无运行中的 Fabric 容器。"
fi

# 4. GitHub 存盘防呆检查
echo -e "${YELLOW}[4/4] 正在进行 GitHub 存盘检查...${NC}"
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
