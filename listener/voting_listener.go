package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"os/exec"
	"encoding/hex"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const ReportExecutedTopic   = "0x9329b41f8c61b0f08b129c86ffbc96416ad3b78ea92188b89598c4667b146c2e"
const VotingContractAddr    = "0xb59e0050ac3449d8a9f7a40670ed86da7d89d5ac"
const ReputationContractAddr2 = "0x0a1b7fadcece391ac1e960d150b8e1ffc06ae162"

func fiscoCallGroup(groupId int, method string, params []interface{}) (json.RawMessage, error) {
	req := rpcReq{JSONRPC: "2.0", Method: method, Params: params, ID: 1}
	body, _ := json.Marshal(req)
	resp, err := http.Post(FiscoRPCURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r rpcResp
	json.Unmarshal(data, &r)
	if r.Error != nil {
		return nil, fmt.Errorf("%s", r.Error.Message)
	}
	return r.Result, nil
}

func getBlockNumberGroup2() (int64, error) {
	result, err := fiscoCallGroup(2, "getBlockNumber", []interface{}{2})
	if err != nil {
		return 0, err
	}
	var hexStr string
	json.Unmarshal(result, &hexStr)
	hexStr = strings.TrimPrefix(hexStr, "0x")
	n := int64(0)
	for _, c := range hexStr {
		n = n*16
		if c >= '0' && c <= '9' { n += int64(c - '0') } else
		if c >= 'a' && c <= 'f' { n += int64(c-'a') + 10 } else
		if c >= 'A' && c <= 'F' { n += int64(c-'A') + 10 }
	}
	return n, nil
}

func getBlockTxsGroup2(blockNum int64) ([]FISCOTx, error) {
	hexBlock := fmt.Sprintf("0x%x", blockNum)
	result, err := fiscoCallGroup(2, "getBlockByNumber", []interface{}{2, hexBlock, true})
	if err != nil {
		return nil, err
	}
	var block FISCOBlock
	json.Unmarshal(result, &block)
	return block.Transactions, nil
}

func getReceiptGroup2(txHash string) (*FISCOReceipt, error) {
	result, err := fiscoCallGroup(2, "getTransactionReceipt", []interface{}{2, txHash})
	if err != nil {
		return nil, err
	}
	var receipt FISCOReceipt
	json.Unmarshal(result, &receipt)
	return &receipt, nil
}

func parseReportExecuted(data string) (reportId string, passed bool) {
	data = strings.TrimPrefix(data, "0x")
	raw, err := hex.DecodeString(data)
	if err != nil || len(raw) < 96 {
		return "", false
	}
	passed = raw[95] == 1
	str := string(raw)
	idx := strings.Index(str, "report")
	if idx >= 0 {
		end := idx
		for end < len(str) && str[end] >= 32 && str[end] < 127 {
			end++
		}
		reportId = str[idx:end]
	}
	return reportId, passed
}

func getTargetCADID(reportId string) string {
	cmd := exec.Command("bash", "console.sh", "2", "call", "CAVoting",
		VotingContractAddr, "getReport", reportId)
	cmd.Dir = Cfg.Fisco.ConsoleDir
	out, _ := cmd.CombinedOutput()
	re := regexp.MustCompile(`did:pki:[^\s,)]+`)
	return re.FindString(string(out))
}

func triggerReputationFromVote(reportId string) {
	caDID := getTargetCADID(reportId)
	if caDID == "" {
		log.Printf("[Vote] 无法获取 targetCADID: %s\n", reportId)
		return
	}
	fmt.Printf("[Vote] 投票通过，对 CA %s 扣分\n", caDID)
	cmd := exec.Command("bash", "console.sh", "call", "CAReputation",
		ReputationContractAddr2, "onCertReported", caDID)
	cmd.Dir = Cfg.Fisco.ConsoleDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Vote] 扣分失败: %v\n", err)
		return
	}
	fmt.Printf("[Vote] 扣分成功: %s\n", string(out))
}

func startGroup2Polling() {
	fmt.Println("[Poller-G2] 开始监听 Group2 ReportExecuted 事件...")
	lastBlock := int64(0)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		current, err := getBlockNumberGroup2()
		if err != nil || current <= lastBlock {
			continue
		}
		for blockNum := lastBlock + 1; blockNum <= current; blockNum++ {
			txs, err := getBlockTxsGroup2(blockNum)
			if err != nil {
				continue
			}
			for _, tx := range txs {
				if !strings.EqualFold(tx.To, VotingContractAddr) {
					continue
				}
				if alreadyProcessed(tx.Hash) {
					continue
				}
				markProcessed(tx.Hash)
				receipt, err := getReceiptGroup2(tx.Hash)
				if err != nil || receipt == nil {
					continue
				}
				for _, l := range receipt.Logs {
					if len(l.Topics) == 0 {
						continue
					}
					if !strings.EqualFold(l.Topics[0], ReportExecutedTopic) {
						continue
					}
					reportId, passed := parseReportExecuted(l.Data)
					fmt.Printf("[Poller-G2] 区块 %d ReportExecuted: %s passed=%v\n",
						blockNum, reportId, passed)
					if passed && reportId != "" {
						go triggerReputationFromVote(reportId)
					}
				}
			}
		}
		lastBlock = current
	}
}
