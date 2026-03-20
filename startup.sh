#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== [Cross-Chain Oracle] 极客全境终极点火序列 v2.0 ===${NC}"

# 1. 斩断时间刺客
echo -e "${YELLOW}[1/5] 正在同步系统时间...${NC}"
sudo timedatectl set-ntp no && sudo date -s "$(curl -sI baidu.com | grep -i '^date:' | cut -d' ' -f2-7)" && sudo timedatectl set-ntp yes
echo -e "${GREEN}时间同步完成！${NC}"

# 2. 唤醒 FISCO BCOS
echo -e "${YELLOW}[2/5] 正在唤醒 FISCO BCOS 底层节点...${NC}"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash start_all.sh
    echo -e "${GREEN}FISCO BCOS 节点已启动！${NC}"
else
    echo -e "${RED}[错误] 未找到 FISCO 节点路径。${NC}"
fi

# 3. 唤醒 Fabric 与智能合约
echo -e "${YELLOW}[3/5] 正在处理 Hyperledger Fabric 网络...${NC}"
cd $HOME/fabric-project/fabric-samples/test-network/
if [ "$(docker ps -aq)" ]; then
    echo "检测到存量 Docker 容器，正在直接唤醒底层网络..."
    docker start $(docker ps -aq) > /dev/null
    echo -e "${GREEN}Fabric 容器唤醒完毕！${NC}"
else
    echo "未检测到容器，执行深度冷启动..."
    ./network.sh up createChannel -c mychannel -s couchdb
    echo "正在将 real-pkicert 智能合约自动部署上链..."
    ./network.sh deployCC -ccn pki -ccp ../real-pkicert -ccl go
    echo -e "${GREEN}Fabric 网络与智能合约冷部署完毕！${NC}"
fi

# 4. 注入 ABI 字典
echo -e "${YELLOW}[4/5] 正在向 FISCO 控制台注入 ABI 字典...${NC}"
mkdir -p $HOME/console/conf/abi
mkdir -p $HOME/console/contracts/sdk/abi
ABI_SRC="$HOME/cross-chain-project/listener/api/CrossChainClient.abi"
if [ -f "$ABI_SRC" ]; then
    cp $ABI_SRC $HOME/console/conf/abi/ 2>/dev/null
    cp $ABI_SRC $HOME/console/contracts/sdk/abi/ 2>/dev/null
    echo -e "${GREEN}ABI 字典注入成功！回写障碍已扫清。${NC}"
else
    echo -e "${RED}[警告] 未找到 ABI 文件。${NC}"
fi

# 5. 点火预言机适配器
echo -e "${YELLOW}[5/5] 正在启动 Fabric 适配器后台服务...${NC}"
pkill -f "adapter.go" 2>/dev/null
sleep 1

if [ -d "$HOME/cross-chain-project/chainlink-adapter" ]; then
    cd $HOME/cross-chain-project/chainlink-adapter
    nohup go run adapter.go > adapter_run.log 2>&1 &
    sleep 2
    if pgrep -f "adapter.go" > /dev/null; then
        echo -e "${GREEN}Fabric 适配器已成功在后台运行 (监听 8081 端口)。${NC}"
    else
        echo -e "${RED}[错误] 适配器启动失败，请检查 adapter_run.log 的报错信息。${NC}"
    fi
fi

echo "----------------------------------------------------------------------"
echo -e "${GREEN}全链路底层服务已全部就位！系统处于实战待命状态。${NC}"
echo "----------------------------------------------------------------------"
