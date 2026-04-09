#!/bin/bash
# start_all.sh - Cross-Chain Oracle 系统启动脚本

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
NC='\033[0m'

PROJECT_DIR=$(cd "$(dirname "$0")"; pwd)
BIN_DIR="$PROJECT_DIR/bin"
LOG_DIR="$PROJECT_DIR/logs"
mkdir -p "$BIN_DIR" "$LOG_DIR"

step() { echo -e "\n${BLUE}▶ $1${NC}"; }
ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }
fail() { echo -e "  ${RED}✗${NC} $1"; }
info() { echo -e "  ${GRAY}·${NC} $1"; }

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Cross-Chain Oracle  ·  系统启动${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ── 1. 时间同步 ────────────────────────────────────────────
step "同步系统时间"
sudo timedatectl set-ntp no && \
sudo date -s "$(curl -sI baidu.com | grep -i '^date:' | cut -d' ' -f2-7)" > /dev/null 2>&1 && \
sudo timedatectl set-ntp yes
ok "$(date '+%Y-%m-%d %H:%M:%S CST')"

# ── 2. IPFS ────────────────────────────────────────────────
step "IPFS 去中心化存储"
if pgrep -x "ipfs" > /dev/null; then
    ok "IPFS 已在运行 (port 5001)"
else
    nohup ipfs daemon > "$LOG_DIR/ipfs.log" 2>&1 &
    sleep 3
    if pgrep -x "ipfs" > /dev/null; then
        ok "IPFS 节点已启动 (port 5001)"
    else
        fail "IPFS 启动失败，请检查 ipfs init 是否已执行"
        exit 1
    fi
fi

# ── 3. FISCO BCOS ──────────────────────────────────────────
step "FISCO BCOS 区块链节点"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash start_all.sh > /dev/null 2>&1
    ok "FISCO BCOS 节点已启动 (4 节点)"
else
    fail "未找到 FISCO 节点路径: ~/fisco/nodes/127.0.0.1"
fi

# ── 4. Hyperledger Fabric ──────────────────────────────────
step "Hyperledger Fabric 网络"
cd $HOME/fabric-project/fabric-samples/test-network/
FABRIC_CONTAINERS=$(docker ps -a -q --filter "name=peer" --filter "name=orderer" --filter "name=couchdb" --filter "name=cli")
if [ -n "$FABRIC_CONTAINERS" ]; then
    docker start $FABRIC_CONTAINERS > /dev/null 2>&1
    ok "Fabric 容器已唤醒 (channel: mychannel, chaincode: pki)"
else
    warn "未检测到容器，执行冷启动..."
    ./network.sh up createChannel -c mychannel -s couchdb > /dev/null 2>&1
    ./network.sh deployCC -ccn pki -ccp ../real-pkicert -ccl go > /dev/null 2>&1
    ok "Fabric 网络冷启动完成"
fi

# ── 5. Chainlink ───────────────────────────────────────────
step "Chainlink 预言机节点"
if [ -d "$HOME/cross-chain-project/chainlink-node" ]; then
    cd $HOME/cross-chain-project/chainlink-node
    docker-compose start > /dev/null 2>&1
    ok "Chainlink 节点与 PostgreSQL 已上线 (port 6688)"
else
    fail "未找到 Chainlink 目录"
fi

# ── 6. 编译微服务 ──────────────────────────────────────────
step "编译跨链微服务"

compile() {
    local name=$1
    local dir=$2
    local files=$3
    cd "$dir"
    if go build -o "$BIN_DIR/$name" $files 2>/dev/null; then
        ok "$name"
    else
        fail "$name 编译失败"
        go build -o "$BIN_DIR/$name" $files
        exit 1
    fi
}

compile "issuer_ui"      "$PROJECT_DIR"                  "issuer_ui.go config.go"
compile "verifier_ui"    "$PROJECT_DIR"                  "verifier_ui.go config.go"
compile "fabric_adapter" "$PROJECT_DIR/chainlink-adapter" "adapter.go config.go"
compile "auto_trigger"   "$PROJECT_DIR/listener"          "auto_trigger.go voting_listener.go config.go"
compile "fisco_writer"   "$PROJECT_DIR/listener"          "fisco_writer.go config.go"

# ── 7. 启动微服务 ──────────────────────────────────────────
step "启动跨链微服务"

start_service() {
    local name=$1
    local run_dir=$2
    cd "$run_dir"
    nohup "$BIN_DIR/$name" > "$LOG_DIR/${name}.log" 2>&1 &
    echo $! > "$LOG_DIR/${name}.pid"
    ok "$name  (PID: $(cat "$LOG_DIR/${name}.pid"))"
}

start_service "fabric_adapter" "$PROJECT_DIR/chainlink-adapter"
start_service "fisco_writer"   "$PROJECT_DIR/listener"
start_service "auto_trigger"   "$PROJECT_DIR/listener"
start_service "issuer_ui"      "$PROJECT_DIR"
start_service "verifier_ui"    "$PROJECT_DIR"

# ── 完成 ───────────────────────────────────────────────────
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  系统已就绪${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  ${GRAY}存证系统${NC}    http://localhost:8889"
echo -e "  ${GRAY}核验系统${NC}    http://localhost:8888"
echo -e "  ${GRAY}预言机${NC}      http://localhost:6688"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
