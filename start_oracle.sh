#!/bin/bash

# 终端输出高亮配色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}[Oracle Startup] 开始执行跨链预言机一键启动序列...${NC}"

PROJECT_DIR="$HOME/cross-chain-project"
CONSOLE_DIR="$HOME/console"

# 1. 终极修复：向控制台强行注入 ABI 字典
echo -e "${YELLOW}[Oracle Startup] 正在向 FISCO BCOS 控制台注入智能合约 ABI 字典...${NC}"
mkdir -p $CONSOLE_DIR/conf/abi
mkdir -p $CONSOLE_DIR/contracts/sdk/abi
cp $PROJECT_DIR/listener/api/CrossChainClient.abi $CONSOLE_DIR/conf/abi/ 2>/dev/null
cp $PROJECT_DIR/listener/api/CrossChainClient.abi $CONSOLE_DIR/contracts/sdk/abi/ 2>/dev/null
echo -e "${GREEN}[Oracle Startup] ABI 字典注入完毕，最后 1% 的回写障碍已扫清。${NC}"

# 2. 战场清理：斩断历史僵尸进程
echo -e "${YELLOW}[Oracle Startup] 扫描并清理历史挂起的预言机后台进程...${NC}"
pkill -f "adapter.go" 2>/dev/null
sleep 1

# 3. 引擎点火：后台静默启动 Fabric 适配器
echo -e "${YELLOW}[Oracle Startup] 正在点火 Fabric 适配器节点...${NC}"
cd $PROJECT_DIR/chainlink-adapter
nohup go run adapter.go > adapter_run.log 2>&1 &
sleep 2

# 4. 存活探针：检查适配器心跳
if pgrep -f "adapter.go" > /dev/null; then
    echo -e "${GREEN}[Oracle Startup] Fabric 适配器已成功在后台运行 (监听端口 8081)。${NC}"
    echo -e "${GREEN}[Oracle Startup] 日志已重定向至: $PROJECT_DIR/chainlink-adapter/adapter_run.log${NC}"
else
    echo "[Fatal Error] Fabric 适配器启动失败，请检查端口是否被占用或查看 adapter_run.log"
    exit 1
fi

echo "----------------------------------------------------------------------"
echo -e "${GREEN}全链路底层服务已就位！你的系统现在处于实战待命状态。${NC}"
echo ""
echo "接下来的战术动作："
echo "1. 切到 FISCO 控制台，打出跨链查证子弹并复制返回的 transaction hash："
echo "   call CrossChainClient 0x77580befe1c74597dfdd626d6cced71807b868d4 requestVerification \"QmTestHash123456789\""
echo ""
echo "2. 在当前终端进入预言机目录，带着哈希值发起截获与回写："
echo "   cd $PROJECT_DIR/listener"
echo "   go run main.go [你的TransactionHash]"
echo "----------------------------------------------------------------------"
