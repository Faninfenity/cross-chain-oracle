#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain] 启动全境唤醒序列 ===${NC}"

# 1. 唤醒 FISCO BCOS 底层链节点
echo -e "${YELLOW}[1/4] 正在唤醒 FISCO BCOS 底层节点...${NC}"
if [ -d "$HOME/nodes/127.0.0.1" ]; then
    bash $HOME/nodes/127.0.0.1/start_all.sh
    echo -e "${GREEN}FISCO BCOS 节点已启动！${NC}"
else
    echo -e "${RED}未找到 FISCO BCOS 节点目录，请确认是否在 ~/nodes/127.0.0.1${NC}"
fi

# 2. 唤醒 Fabric 网络 (老范，如果你的Fabric路径不同，请修改这里)
echo -e "${YELLOW}[2/4] 正在唤醒 Hyperledger Fabric 测试网络...${NC}"
FABRIC_DIR="$HOME/fabric-samples/test-network"
if [ -d "$FABRIC_DIR" ]; then
    cd $FABRIC_DIR
    ./network.sh up createChannel -c mychannel -s couchdb
    echo -e "${GREEN}Fabric 网络启动指令已下达！${NC}"
else
    echo -e "${YELLOW}提示: 未在默认路径 ($FABRIC_DIR) 找到 Fabric，如果你已通过其他方式启动或路径不同，请忽略此提示。${NC}"
fi

# 3. 注入控制台 ABI 字典 (解决最后的 1% 报错)
echo -e "${YELLOW}[3/4] 正在向 FISCO 控制台注入 ABI 字典...${NC}"
mkdir -p $HOME/console/conf/abi
mkdir -p $HOME/console/contracts/sdk/abi
cp $HOME/cross-chain-project/listener/api/CrossChainClient.abi $HOME/console/conf/abi/ 2>/dev/null
cp $HOME/cross-chain-project/listener/api/CrossChainClient.abi $HOME/console/contracts/sdk/abi/ 2>/dev/null
echo -e "${GREEN}ABI 字典注入完毕！${NC}"

# 4. 点火预言机适配器
echo -e "${YELLOW}[4/4] 正在清理并重启 Fabric 适配器后台服务...${NC}"
pkill -f "adapter.go" 2>/dev/null
sleep 1

cd $HOME/cross-chain-project/chainlink-adapter
nohup go run adapter.go > adapter_run.log 2>&1 &
sleep 2

if pgrep -f "adapter.go" > /dev/null; then
    echo -e "${GREEN}Fabric 适配器已成功在后台运行 (监听 8081 端口)。${NC}"
else
    echo -e "${RED}[Fatal Error] 适配器启动失败，请检查端口占用！${NC}"
fi

echo -e "${GREEN}=== 全链路环境已全部就绪！ ===${NC}"
