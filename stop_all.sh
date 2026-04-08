#!/bin/bash
# stop_all.sh - Cross-Chain Oracle 系统停止脚本

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
NC='\033[0m'

PROJECT_DIR=$(cd "$(dirname "$0")"; pwd)
LOG_DIR="$PROJECT_DIR/logs"

step() { echo -e "\n${BLUE}▶ $1${NC}"; }
ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }
fail() { echo -e "  ${RED}✗${NC} $1"; }
info() { echo -e "  ${GRAY}·${NC} $1"; }

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Cross-Chain Oracle  ·  系统停止${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ── 1. 停止微服务 ──────────────────────────────────────────
step "停止跨链微服务"
SERVICES=("issuer_ui" "verifier_ui" "auto_trigger" "fisco_writer" "fabric_adapter")
for service in "${SERVICES[@]}"; do
    pid_file="$LOG_DIR/${service}.pid"
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            kill $pid
            ok "$service (PID: $pid)"
        else
            info "$service 进程已不存在"
        fi
        rm -f "$pid_file"
    else
        pkill -f "$service" > /dev/null 2>&1
        info "$service (pid 文件不存在，已尝试 pkill)"
    fi
done

# ── 2. IPFS ────────────────────────────────────────────────
step "停止 IPFS"
if pgrep -x "ipfs" > /dev/null; then
    ipfs shutdown > /dev/null 2>&1
    ok "IPFS 已安全停止"
else
    info "IPFS 未在运行"
fi

# ── 3. FISCO BCOS ──────────────────────────────────────────
step "停止 FISCO BCOS"
if [ -d "$HOME/fisco/nodes/127.0.0.1" ]; then
    cd $HOME/fisco/nodes/127.0.0.1/ && bash stop_all.sh > /dev/null 2>&1
    ok "FISCO BCOS 已停止"
else
    warn "未找到 FISCO 节点路径"
fi

# ── 4. Chainlink ───────────────────────────────────────────
step "停止 Chainlink"
if [ -d "$HOME/cross-chain-project/chainlink-node" ]; then
    cd $HOME/cross-chain-project/chainlink-node
    docker-compose stop > /dev/null 2>&1
    ok "Chainlink 节点与 PostgreSQL 已停止"
else
    warn "未找到 Chainlink 目录"
fi

# ── 5. Fabric ──────────────────────────────────────────────
step "休眠 Fabric 容器"
remaining=$(docker ps -q)
if [ -n "$remaining" ]; then
    docker stop $remaining > /dev/null 2>&1
    ok "所有 Docker 容器已停止"
else
    info "无运行中的容器"
fi

# ── 6. Git 存盘检查 ────────────────────────────────────────
step "代码存盘检查"
cd "$PROJECT_DIR"
if [[ -n $(git status -s) ]]; then
    warn "发现未提交的修改："
    git status -s | while read line; do info "$line"; done
    echo ""
    echo -e "  ${YELLOW}运行 git add -A && git commit -m 'update' && git push 后再关机${NC}"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}  服务已全部停止，但请先提交代码再关机${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
else
    ok "代码已全部提交"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  系统已完全停止，可以安全关机${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
fi
