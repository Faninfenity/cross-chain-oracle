#!/bin/bash
# benchmark2.sh - 完整性能压测（真并发）

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
ORG1_TLS=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
DID="did:pki:org1:bf256c573a917b2f"

# OR策略只需一个peer
PEER_ARGS="--peerAddresses localhost:7051 --tlsRootCertFiles $ORG1_TLS"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  系统性能压测 v2（真并发测试）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ── 场景一：真并发写入 TPS（多梯度）─────────────────────
echo -e "\n${YELLOW}▶ 场景一：真并发存证 TPS（OR策略，单节点背书）${NC}"
echo -e "  并发数 | 总笔数 | 耗时(ms) | TPS | 成功率"
echo -e "  -------|--------|----------|-----|-------"

for CONCURRENCY in 10 20 50 100; do
    TOTAL=$((CONCURRENCY * 2))
    SUCCESS=0
    TMPDIR=$(mktemp -d)

    START=$(date +%s%3N)
    for i in $(seq 1 $TOTAL); do
        CID="QmV2Test${CONCURRENCY}_$(date +%s%N)_${i}_$$"
        peer chaincode invoke \
            -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
            --tls --cafile $ORDERER_CA \
            -C mychannel -n pki \
            $PEER_ARGS \
            -c "{\"function\":\"CreateCert\",\"Args\":[\"$CID\",\"$DID\",\"$CID\"]}" \
            > $TMPDIR/tx_${i}.log 2>&1 &
    done
    wait
    END=$(date +%s%3N)

    for f in $TMPDIR/tx_*.log; do
        grep -q "successful" $f && SUCCESS=$((SUCCESS+1))
    done
    rm -rf $TMPDIR

    ELAPSED=$((END - START))
    TPS=$(echo "scale=2; $TOTAL * 1000 / $ELAPSED" | bc)
    RATE="${SUCCESS}/${TOTAL}"
    echo -e "  ${CONCURRENCY}      | ${TOTAL}     | ${ELAPSED}      | ${TPS} | ${RATE}"
    sleep 2
done

# ── 场景二：MAJORITY策略 vs OR策略 TPS 对比 ──────────────
echo -e "\n${YELLOW}▶ 场景二：背书策略对比（50并发，100笔）${NC}"

# OR策略
SUCCESS=0
TMPDIR=$(mktemp -d)
START=$(date +%s%3N)
for i in $(seq 1 100); do
    CID="QmOR_$(date +%s%N)_${i}"
    peer chaincode invoke \
        -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile $ORDERER_CA \
        -C mychannel -n pki \
        $PEER_ARGS \
        -c "{\"function\":\"CreateCert\",\"Args\":[\"$CID\",\"$DID\",\"$CID\"]}" \
        > $TMPDIR/tx_${i}.log 2>&1 &
done
wait
END=$(date +%s%3N)
for f in $TMPDIR/tx_*.log; do grep -q "successful" $f && SUCCESS=$((SUCCESS+1)); done
rm -rf $TMPDIR
OR_TPS=$(echo "scale=2; 100 * 1000 / $((END-START))" | bc)
OR_TIME=$((END-START))
echo -e "  OR策略（任意1个组织）：${OR_TPS} TPS，耗时 ${OR_TIME}ms，成功 ${SUCCESS}/100"

sleep 3

# MAJORITY策略（6个peer）
MAJORITY_ARGS=""
for ORG_NUM in 1 2 3 5 6 7; do
    if [ $ORG_NUM -eq 1 ]; then PORT=7051
    elif [ $ORG_NUM -eq 2 ]; then PORT=9051
    else PORT=$((11051 + (ORG_NUM - 3) * 2000)); fi
    DOMAIN="org${ORG_NUM}.example.com"
    TLS="${PWD}/organizations/peerOrganizations/${DOMAIN}/peers/peer0.${DOMAIN}/tls/ca.crt"
    MAJORITY_ARGS="$MAJORITY_ARGS --peerAddresses localhost:$PORT --tlsRootCertFiles $TLS"
done

SUCCESS=0
TMPDIR=$(mktemp -d)
START=$(date +%s%3N)
for i in $(seq 1 100); do
    CID="QmMAJ_$(date +%s%N)_${i}"
    peer chaincode invoke \
        -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
        --tls --cafile $ORDERER_CA \
        -C mychannel -n pki \
        $MAJORITY_ARGS \
        -c "{\"function\":\"CreateCert\",\"Args\":[\"$CID\",\"$DID\",\"$CID\"]}" \
        > $TMPDIR/tx_${i}.log 2>&1 &
done
wait
END=$(date +%s%3N)
for f in $TMPDIR/tx_*.log; do grep -q "successful" $f && SUCCESS=$((SUCCESS+1)); done
rm -rf $TMPDIR
MAJ_TPS=$(echo "scale=2; 100 * 1000 / $((END-START))" | bc)
MAJ_TIME=$((END-START))
echo -e "  MAJORITY策略（6组织背书）：${MAJ_TPS} TPS，耗时 ${MAJ_TIME}ms，成功 ${SUCCESS}/100"

# ── 场景三：查询延迟（多次采样）────────────────────────
echo -e "\n${YELLOW}▶ 场景三：查询延迟精确测量（100次采样）${NC}"
KNOWN_CID=$(peer chaincode query -C mychannel -n pki \
    -c '{"Args":["GetAllCerts"]}' 2>&1 | python3 -c "
import json,sys
certs=json.load(sys.stdin)
certs.sort(key=lambda x: x.get('issuedAt',''), reverse=True)
print(certs[0]['certID'])
" 2>/dev/null)

TOTAL_LAT=0
MIN_LAT=99999
MAX_LAT=0
for i in $(seq 1 100); do
    T1=$(date +%s%3N)
    peer chaincode query -C mychannel -n pki \
        -c "{\"Args\":[\"QueryCert\",\"$KNOWN_CID\"]}" > /dev/null 2>&1
    T2=$(date +%s%3N)
    LAT=$((T2-T1))
    TOTAL_LAT=$((TOTAL_LAT+LAT))
    [ $LAT -lt $MIN_LAT ] && MIN_LAT=$LAT
    [ $LAT -gt $MAX_LAT ] && MAX_LAT=$LAT
done
AVG_LAT=$(echo "scale=1; $TOTAL_LAT / 100" | bc)
echo -e "  最小延迟: ${MIN_LAT}ms"
echo -e "  最大延迟: ${MAX_LAT}ms"
echo -e "  平均延迟: ${AVG_LAT}ms"

# ── 场景四：全量查询扩展性 ───────────────────────────────
echo -e "\n${YELLOW}▶ 场景四：全量查询扩展性（不同数据量）${NC}"
for label in "当前"; do
    T1=$(date +%s%3N)
    COUNT=$(peer chaincode query -C mychannel -n pki \
        -c '{"Args":["GetAllCerts"]}' 2>&1 | python3 -c "
import json,sys
certs=json.load(sys.stdin)
print(len(certs))
" 2>/dev/null)
    T2=$(date +%s%3N)
    echo -e "  数据量 ${COUNT} 条：耗时 $((T2-T1))ms"
done

# ── 汇总 ─────────────────────────────────────────────────
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  压测完成${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  OR策略峰值TPS：    ${OR_TPS}"
echo -e "  MAJORITY策略TPS：  ${MAJ_TPS}"
echo -e "  查询平均延迟：     ${AVG_LAT}ms"
echo -e "  查询最小延迟：     ${MIN_LAT}ms"
echo -e "  查询最大延迟：     ${MAX_LAT}ms"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
