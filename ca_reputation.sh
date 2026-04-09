#!/bin/bash
# ca_reputation.sh - CA 信誉分管理工具
# 用法: bash ca_reputation.sh <命令> <DID> [分数] [原因]

CONSOLE_DIR="/home/fan/console"
CONTRACT="CAReputation"
ADDR="0x0a1b7fadcece391ac1e960d150b8e1ffc06ae162"

call() {
    cd $CONSOLE_DIR
    bash console.sh call $CONTRACT $ADDR "$@" 2>&1
}

case "$1" in
    query)
        # 查询信誉分
        # 用法: bash ca_reputation.sh query <DID>
        echo "查询 CA 信誉分: $2"
        call getCAInfo "$2" | grep "Return values"
        ;;
    set)
        # 设置信誉分
        # 用法: bash ca_reputation.sh set <DID> <分数> [原因]
        REASON=${4:-"手动调整"}
        echo "设置 $2 信誉分为 $3 分，原因: $REASON"
        call setScore "$2" $3 "$REASON" | grep -E "successfully|Error|Event"
        ;;
    revoke)
        # 模拟证书吊销扣分（-5分）
        # 用法: bash ca_reputation.sh revoke <DID>
        echo "模拟证书吊销，$2 扣 5 分"
        call onCertRevoked "$2" | grep -E "successfully|Error|Event"
        ;;
    report)
        # 举报 CA（-10分）
        # 用法: bash ca_reputation.sh report <DID>
        echo "举报 CA，$2 扣 10 分"
        call onCertReported "$2" | grep -E "successfully|Error|Event"
        ;;
    audit)
        # 监管审查通过（+5分）
        # 用法: bash ca_reputation.sh audit <DID>
        echo "监管审查通过，$2 加 5 分"
        call onAuditPassed "$2" | grep -E "successfully|Error|Event"
        ;;
    register)
        # 注册新 CA
        # 用法: bash ca_reputation.sh register <DID>
        echo "注册 CA: $2"
        call registerCA "$2" | grep -E "successfully|Error|Event"
        ;;
    trusted)
        # 快捷：设置为 TRUSTED（95分）
        call setScore "$2" 95 "恢复信任" | grep -E "successfully|Event"
        echo "$2 已恢复为 TRUSTED（95分）"
        ;;
    warning)
        # 快捷：设置为 WARNING（45分）
        call setScore "$2" 45 "信誉预警" | grep -E "successfully|Event"
        echo "$2 已设置为 WARNING（45分）"
        ;;
    blocked)
        # 快捷：设置为 BLOCKED（20分）
        call setScore "$2" 20 "信誉封禁" | grep -E "successfully|Event"
        echo "$2 已设置为 BLOCKED（20分）"
        ;;
    *)
        echo "CA 信誉分管理工具"
        echo ""
        echo "用法: bash ca_reputation.sh <命令> <DID> [参数]"
        echo ""
        echo "命令:"
        echo "  query   <DID>              查询信誉分"
        echo "  set     <DID> <分> [原因]  设置信誉分"
        echo "  revoke  <DID>              模拟吊销扣5分"
        echo "  report  <DID>              举报扣10分"
        echo "  audit   <DID>              审查通过加5分"
        echo "  register <DID>             注册新CA"
        echo "  trusted  <DID>             快捷设为TRUSTED(95分)"
        echo "  warning  <DID>             快捷设为WARNING(45分)"
        echo "  blocked  <DID>             快捷设为BLOCKED(20分)"
        echo ""
        echo "已知 DID:"
        echo "  Org1: did:pki:org1:bf256c573a917b2f"
        echo "  Org3: did:pki:org3:7940d92f899e2348"
        ;;
esac
