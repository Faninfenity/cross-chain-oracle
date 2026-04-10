#!/bin/bash
# benchmark.sh - 系统性能压测脚本

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

cd ~/fabric-project/fabric-samples/test-network
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp

ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# 动态构建 peer 参数
# OR策略下只需一个peer背书
PEER_ARGS="--peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  系统性能压测${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ── 场景一：存证 TPS ──────────────────────────────────────
echo -e "\n${YELLOW}▶ 场景一：存证吞吐量测试（连续写入20笔）${NC}"

DID="did:pki:org1:bf256c573a917b2f"
SUCCESS=0
FAIL=0
START=$(date +%s%3N)

for i in $(seq 1 20); do
    CID="QmBenchmark$(date +%s%N)${i}"
    OUT=$(peer chaincode invoke \
        -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile $ORDERER_CA \
        -C mychannel -n pki \
        $PEER_ARGS \
        -c "{\"function\":\"CreateCert\",\"Args\":[\"$CID\",\"$DID\",\"$CID\"]}" 2>&1)
    if echo "$OUT" | grep -q "successful"; then
        SUCCESS=$((SUCCESS + 1))
    else
        FAIL=$((FAIL + 1))
    fi
    echo -n "."
done

END=$(date +%s%3N)
ELAPSED=$(( (END - START) ))
TPS=$(echo "scale=2; 20 * 1000 / $ELAPSED" | bc)

echo ""
echo -e "  成功: ${SUCCESS}/20  失败: ${FAIL}/20"
echo -e "  总耗时: ${ELAPSED}ms"
echo -e "  ${GREEN}存证 TPS: ${TPS}${NC}"

# ── 场景二：查询延迟 ─────────────────────────────────────
echo -e "\n${YELLOW}▶ 场景二：证书查询延迟测试（连续查询20次）${NC}"

# 取一个已知存在的 CID
KNOWN_CID=$(peer chaincode query -C mychannel -n pki \
    -c '{"Args":["GetAllCerts"]}' 2>&1 | python3 -c "
import json,sys
certs=json.load(sys.stdin)
print(certs[0]['certID'])
" 2>/dev/null)

TOTAL_LATENCY=0
for i in $(seq 1 20); do
    T_START=$(date +%s%3N)
    peer chaincode query -C mychannel -n pki \
        -c "{\"Args\":[\"QueryCert\",\"$KNOWN_CID\"]}" > /dev/null 2>&1
    T_END=$(date +%s%3N)
    LATENCY=$((T_END - T_START))
    TOTAL_LATENCY=$((TOTAL_LATENCY + LATENCY))
    echo -n "."
done

AVG_LATENCY=$(echo "scale=1; $TOTAL_LATENCY / 20" | bc)
echo ""
echo -e "  ${GREEN}平均查询延迟: ${AVG_LATENCY}ms${NC}"

# ── 场景三：GetAllCerts 全量查询 ─────────────────────────
echo -e "\n${YELLOW}▶ 场景三：全量台账查询延迟${NC}"

T_START=$(date +%s%3N)
COUNT=$(peer chaincode query -C mychannel -n pki \
    -c '{"Args":["GetAllCerts"]}' 2>&1 | python3 -c "
import json,sys
certs=json.load(sys.stdin)
print(len(certs))
" 2>/dev/null)
T_END=$(date +%s%3N)
FULL_LATENCY=$((T_END - T_START))

echo -e "  账本记录数: ${COUNT} 条"
echo -e "  ${GREEN}全量查询耗时: ${FULL_LATENCY}ms${NC}"

# ── 场景四：verifier_ui API 响应时间 ─────────────────────
echo -e "\n${YELLOW}▶ 场景四：核验 API 端到端响应时间（含CA信誉查询）${NC}"

TOTAL_API=0
for i in $(seq 1 10); do
    T_START=$(date +%s%3N)
    curl -s "http://localhost:8888/api/fabric-query?id=$KNOWN_CID" > /dev/null
    T_END=$(date +%s%3N)
    TOTAL_API=$((TOTAL_API + T_END - T_START))
    echo -n "."
done

AVG_API=$(echo "scale=1; $TOTAL_API / 10" | bc)
echo ""
echo -e "  ${GREEN}平均 API 响应时间: ${AVG_API}ms（含 Fabric 查询 + CA 信誉查询）${NC}"

# ── 汇总 ─────────────────────────────────────────────────
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  压测结果汇总${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  存证吞吐量:        ${TPS} TPS"
echo -e "  单证书查询延迟:    ${AVG_LATENCY} ms"
echo -e "  全量台账查询:      ${FULL_LATENCY} ms（${COUNT}条记录）"
echo -e "  核验API响应时间:   ${AVG_API} ms"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"


# ── 场景五：并发存证测试 ──────────────────────────────────
echo -e "\n${YELLOW}▶ 场景五：并发存证测试${NC}"

cd ~/fabric-project/fabric-samples/test-network
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp

ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
DID="did:pki:org1:bf256c573a917b2f"

PEER_ARGS=""
for ORG_NUM in 1 2 3 5 6 7; do
    if [ $ORG_NUM -eq 1 ]; then PORT=7051
    elif [ $ORG_NUM -eq 2 ]; then PORT=9051
    else PORT=$((11051 + (ORG_NUM - 3) * 2000)); fi
    DOMAIN="org${ORG_NUM}.example.com"
    TLS="${PWD}/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt"
    if docker ps | grep -q "peer0.${DOMAIN}"; then
        PEER_ARGS="$PEER_ARGS --peerAddresses localhost:$PORT --tlsRootCertFiles $TLS"
    fi
done

for CONCURRENCY in 5 10 20; do
    echo -n "  并发${CONCURRENCY}: "
    START=$(date +%s%3N)
    PIDS=""
    SUCCESS=0

    for i in $(seq 1 $CONCURRENCY); do
        CID="QmConcurrent${CONCURRENCY}_$(date +%s%N)_${i}"
        peer chaincode invoke \
            -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
            --tls --cafile $ORDERER_CA \
            -C mychannel -n pki \
            $PEER_ARGS \
            -c "{\"function\":\"CreateCert\",\"Args\":[\"$CID\",\"$DID\",\"$CID\"]}" \
            > /tmp/bench_${i}.log 2>&1 &
        PIDS="$PIDS $!"
    done

    # 等待所有并发请求完成
    for PID in $PIDS; do
        wait $PID
        if grep -q "successful" /tmp/bench_*.log 2>/dev/null; then
            SUCCESS=$((SUCCESS + 1))
        fi
    done

    END=$(date +%s%3N)
    ELAPSED=$((END - START))
    TPS=$(echo "scale=2; $CONCURRENCY * 1000 / $ELAPSED" | bc)
    echo "耗时 ${ELAPSED}ms，TPS: ${TPS}，成功率: ${SUCCESS}/${CONCURRENCY}"
    rm -f /tmp/bench_*.log
    sleep 3
done
