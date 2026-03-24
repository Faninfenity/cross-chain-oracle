#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain Oracle] 极客全境终极点火序列 v6.0 ===${NC}"

# 0. 环境变量与目录准备
PROJECT_DIR=$(cd "$(dirname "$0")"; pwd)
BIN_DIR="$PROJECT_DIR/bin"
LOG_DIR="$PROJECT_DIR/logs"
mkdir -p "$BIN_DIR" "$LOG_DIR"

# 1. 斩断时间刺客
echo -e "${YELLOW}[1/7] 正在同步系统时间...${NC}"
sudo timedatectl set-ntp no && sudo date -s "$(curl -sI baidu.com | grep -i '^date:' | cut -d' ' -f2-7)" && sudo timedatectl set-ntp yes
echo -e "${GREEN}时间同步完成！${NC}"

# 2. 唤醒 IPFS 星际文件系统
echo -e "${YELLOW}[2/7] 正在唤醒 IPFS 去中心化存储节点...${NC}"
if pgrep -x "ipfs" > /dev/null; then
    echo -e "${GREEN}IPFS 节点已在运行中。${NC}"
else
    nohup ipfs daemon > "$LOG_DIR/ipfs.log" 2>&1 &
    # 给 IPFS 几秒钟的初始化时间
    sleep 3 
    if pgrep -x "ipfs" > /dev/null; then
        echo -e "${GREEN}IPFS 守护进程已成功拉起 (API: 5001)。${NC}"
    else
        echo -e "${RED}[错误] IPFS 启动失败，请检查是否已执行 ipfs init。${NC}"
        exit 1
    fi
fi

# 3. 唤醒 FISCO BCOS
echo -e "${YELLOW}[3/7] 正在唤醒 FISCO BCOS 底层节点...${NC}"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash start_all.sh
    echo -e "${GREEN}FISCO BCOS 节点已启动！${NC}"
else
    echo -e "${RED}[错误] 未找到 FISCO 节点路径。${NC}"
fi

# 4. 唤醒 Fabric 与智能合约
echo -e "${YELLOW}[4/7] 正在处理 Hyperledger Fabric 网络...${NC}"
cd $HOME/fabric-project/fabric-samples/test-network/
FABRIC_CONTAINERS=$(docker ps -a -q --filter "name=peer" --filter "name=orderer" --filter "name=couchdb" --filter "name=cli")
if [ -n "$FABRIC_CONTAINERS" ]; then
    echo "检测到存量 Fabric 容器，正在直接唤醒底层网络..."
    docker start $FABRIC_CONTAINERS > /dev/null
    echo -e "${GREEN}Fabric 容器唤醒完毕！${NC}"
else
    echo "未检测到容器，执行深度冷启动..."
    ./network.sh up createChannel -c mychannel -s couchdb
    echo "正在将 real-pkicert 智能合约自动部署上链..."
    ./network.sh deployCC -ccn pki -ccp ../real-pkicert -ccl go
    echo -e "${GREEN}Fabric 网络与智能合约冷部署完毕！${NC}"
fi

# 5. 唤醒 Chainlink 去中心化预言机节点
echo -e "${YELLOW}[5/7] 正在启动 Chainlink 预言机舰队...${NC}"
if [ -d "$HOME/cross-chain-project/chainlink-node" ]; then
    cd $HOME/cross-chain-project/chainlink-node
    docker-compose start > /dev/null 2>&1
    echo -e "${GREEN}Chainlink 节点与 Postgres 数据库已上线！${NC}"
else
    echo -e "${RED}[错误] 未发现 Chainlink 阵地目录。${NC}"
fi

# 6. 预编译跨链核心微服务
echo -e "${YELLOW}[6/7] 正在预编译跨链五大核心微服务...${NC}"
cd "$PROJECT_DIR"
go build -o "$BIN_DIR/issuer_ui" issuer_ui.go || { echo -e "${RED}源头存证大屏编译失败${NC}"; exit 1; }
go build -o "$BIN_DIR/verifier_ui" verifier_ui.go || { echo -e "${RED}查证大屏编译失败${NC}"; exit 1; }

cd "$PROJECT_DIR/chainlink-adapter"
go build -o "$BIN_DIR/fabric_adapter" adapter.go || { echo -e "${RED}Fabric适配器编译失败${NC}"; exit 1; }

cd "$PROJECT_DIR/listener"
go build -o "$BIN_DIR/auto_trigger" auto_trigger.go || { echo -e "${RED}传达室编译失败${NC}"; exit 1; }
go build -o "$BIN_DIR/fisco_writer" fisco_writer.go || { echo -e "${RED}回写中枢编译失败${NC}"; exit 1; }
echo -e "${GREEN}核心微服务编译完毕，准备入列。${NC}"

# 7. 点火全套微服务后台
echo -e "${YELLOW}[7/7] 正在后台启动全套跨链微服务...${NC}"
cd "$PROJECT_DIR"

start_service() {
    local name=$1
    local cmd=$2
    nohup $cmd > "$LOG_DIR/${name}.log" 2>&1 &
    echo $! > "$LOG_DIR/${name}.pid"
    echo "  -> $name 已启动 (PID: $(cat "$LOG_DIR/${name}.pid"))"
}

start_service "fabric_adapter" "$BIN_DIR/fabric_adapter"
start_service "fisco_writer" "$BIN_DIR/fisco_writer"
start_service "auto_trigger" "$BIN_DIR/auto_trigger"
start_service "issuer_ui" "$BIN_DIR/issuer_ui"
start_service "verifier_ui" "$BIN_DIR/verifier_ui"

echo "----------------------------------------------------------------------"
echo -e "${GREEN}全链路底层服务 (包含 IPFS) 已全部就位！系统处于实战待命状态。${NC}"
echo -e "${YELLOW}源头存证入口: http://localhost:8889${NC}"
echo -e "${YELLOW}跨链查证入口: http://localhost:8888${NC}"
echo -e "${YELLOW}预言机控制台: http://localhost:6688${NC}"
echo "----------------------------------------------------------------------"
