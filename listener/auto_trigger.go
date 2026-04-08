// listener/auto_trigger.go
// 主动轮询 FISCO BCOS 区块，捕获 CrossDomainRequested 事件，自动触发 Chainlink Webhook
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ── 常量 ──────────────────────────────────────────────────

// CrossDomainRequested(string,string,string,string) 的 keccak256 topic
const CrossDomainRequestedTopic = "0x90476299232d4c86db673dc29ec3678d4e6f5bf3228784745a048cfb5b2b1dfa"

const FiscoRPCURL = "http://127.0.0.1:8545"

// ── FISCO RPC ─────────────────────────────────────────────

type rpcReq struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type rpcResp struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func fiscoCall(method string, params []interface{}) (json.RawMessage, error) {
	body, _ := json.Marshal(rpcReq{JSONRPC: "2.0", Method: method, Params: params, ID: 1})
	resp, err := http.Post(FiscoRPCURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r rpcResp
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, fmt.Errorf("RPC 错误: %s", r.Error.Message)
	}
	return r.Result, nil
}

func getBlockNumber() (int64, error) {
	result, err := fiscoCall("getBlockNumber", []interface{}{1})
	if err != nil {
		return 0, err
	}
	var hexStr string
	json.Unmarshal(result, &hexStr)
	hexStr = strings.TrimPrefix(hexStr, "0x")
	n := int64(0)
	for _, c := range hexStr {
		n = n*16
		if c >= '0' && c <= '9' {
			n += int64(c - '0')
		} else if c >= 'a' && c <= 'f' {
			n += int64(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			n += int64(c-'A') + 10
		}
	}
	return n, nil
}

type FISCOTx struct {
	Hash string `json:"hash"`
	To   string `json:"to"`
}

type FISCOBlock struct {
	Transactions []FISCOTx `json:"transactions"`
}

func getBlockTxs(blockNum int64) ([]FISCOTx, error) {
	hexBlock := fmt.Sprintf("0x%x", blockNum)
	result, err := fiscoCall("getBlockByNumber", []interface{}{1, hexBlock, true})
	if err != nil {
		return nil, err
	}
	var block FISCOBlock
	json.Unmarshal(result, &block)
	return block.Transactions, nil
}

type FISCOLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type FISCOReceipt struct {
	Logs   []FISCOLog `json:"logs"`
	Status string     `json:"status"`
}

func getReceipt(txHash string) (*FISCOReceipt, error) {
	result, err := fiscoCall("getTransactionReceipt", []interface{}{1, txHash})
	if err != nil {
		return nil, err
	}
	var receipt FISCOReceipt
	if err := json.Unmarshal(result, &receipt); err != nil {
		return nil, err
	}
	return &receipt, nil
}

// ── ABI 解析 ──────────────────────────────────────────────
// CrossDomainRequested(string reqId, string sourceDID, string targetDomain, string dataFingerprint)
// 所有参数非 indexed，全在 data 里，ABI 编码：4个offset(各32字节) + 4个(length+data)

type CrossDomainEvent struct {
	ReqId           string
	SourceDID       string
	TargetDomain    string
	DataFingerprint string
}

func parseABIStrings(data string) []string {
	data = strings.TrimPrefix(data, "0x")
	raw, err := hex.DecodeString(data)
	if err != nil || len(raw) < 128 {
		return nil
	}

	readUint256 := func(offset int) int {
		if offset+32 > len(raw) {
			return 0
		}
		v := 0
		for _, b := range raw[offset : offset+32] {
			v = v*256 + int(b)
		}
		return v
	}

	readString := func(offset int) string {
		if offset+32 > len(raw) {
			return ""
		}
		length := readUint256(offset)
		if length == 0 || offset+32+length > len(raw) {
			return ""
		}
		return string(raw[offset+32 : offset+32+length])
	}

	// 读4个偏移量
	results := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		offset := readUint256(i * 32)
		if offset > 0 {
			s := readString(offset)
			results = append(results, s)
		}
	}
	return results
}

func parseCrossDomainEvent(data string) *CrossDomainEvent {
	parts := parseABIStrings(data)
	if len(parts) < 4 {
		// 降级：直接扫描 Qm 开头的 CID
		raw, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		if err != nil {
			return nil
		}
		str := string(raw)
		idx := strings.Index(str, "Qm")
		if idx >= 0 && len(str) >= idx+46 {
			cid := str[idx : idx+46]
			return &CrossDomainEvent{
				ReqId:           cid,
				DataFingerprint: cid,
			}
		}
		return nil
	}
	return &CrossDomainEvent{
		ReqId:           parts[0],
		SourceDID:       parts[1],
		TargetDomain:    parts[2],
		DataFingerprint: parts[3],
	}
}

// ── 去重缓存 ──────────────────────────────────────────────

var processedTxs = make(map[string]bool)

func alreadyProcessed(txHash string) bool { return processedTxs[txHash] }

func markProcessed(txHash string) {
	processedTxs[txHash] = true
	if len(processedTxs) > 1000 {
		count := 0
		for k := range processedTxs {
			delete(processedTxs, k)
			if count++; count >= 500 {
				break
			}
		}
	}
}

// ── Chainlink Webhook ─────────────────────────────────────

func triggerChainlinkWebhook(reqId string, certHash string) {
	fmt.Printf("[Oracle] 正在鉴权 Chainlink...\n")
	loginData, _ := json.Marshal(map[string]string{
		"email":    Cfg.Chainlink.Email,
		"password": Cfg.Chainlink.Password,
	})
	client := &http.Client{Timeout: 10 * time.Second}
	loginReq, _ := http.NewRequest("POST", Cfg.Chainlink.SessionURL(), bytes.NewBuffer(loginData))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Printf("[Error] Chainlink 连接失败: %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != 200 {
		log.Printf("[Error] Chainlink 登录失败: %d\n", loginResp.StatusCode)
		return
	}
	var cookie string
	for _, c := range loginResp.Cookies() {
		cookie += c.Name + "=" + c.Value + ";"
	}
	fmt.Printf("[Oracle] 鉴权成功\n")

	payload, _ := json.Marshal(map[string]string{"reqId": reqId, "certHash": certHash})
	req, _ := http.NewRequest("POST", Cfg.Chainlink.WebhookURL(), bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Error] Webhook 触发失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[Oracle] Chainlink Webhook 已触发，HTTP %d\n", resp.StatusCode)
}

// ── 主轮询 ────────────────────────────────────────────────

func startPolling() {
	fmt.Println("[Poller] 开始轮询 FISCO BCOS 区块...")

	lastBlock, err := getBlockNumber()
	if err != nil {
		log.Printf("[Poller] 获取初始区块号失败: %v\n", err)
		lastBlock = 0
	}
	fmt.Printf("[Poller] 起始区块: %d (0x%x)\n", lastBlock, lastBlock)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		current, err := getBlockNumber()
		if err != nil {
			log.Printf("[Poller] 获取区块号失败: %v\n", err)
			continue
		}
		if current <= lastBlock {
			continue
		}

		for blockNum := lastBlock + 1; blockNum <= current; blockNum++ {
			txs, err := getBlockTxs(blockNum)
			if err != nil {
				log.Printf("[Poller] 获取区块 %d 交易失败: %v\n", blockNum, err)
				continue
			}

			for _, tx := range txs {
				// 只处理发给目标合约的交易
				if !strings.EqualFold(tx.To, Cfg.Fisco.ContractAddr) {
					continue
				}
				if alreadyProcessed(tx.Hash) {
					continue
				}
				markProcessed(tx.Hash)

				// 查 Receipt 看事件
				receipt, err := getReceipt(tx.Hash)
				if err != nil || receipt == nil {
					continue
				}

				for _, l := range receipt.Logs {
					if len(l.Topics) == 0 {
						continue
					}
					// 过滤 CrossDomainRequested 事件
					if !strings.EqualFold(l.Topics[0], CrossDomainRequestedTopic) {
						continue
					}

					fmt.Printf("\n[Poller] 区块 %d 捕获到 CrossDomainRequested 事件！\n", blockNum)
					fmt.Printf("[Poller] TxHash: %s\n", tx.Hash)

					event := parseCrossDomainEvent(l.Data)
					if event == nil || event.DataFingerprint == "" {
						fmt.Printf("[Poller] 事件解析失败，跳过\n")
						continue
					}

					fmt.Printf("[Poller] ReqID: %s\n", event.ReqId)
					fmt.Printf("[Poller] CID:   %s\n", event.DataFingerprint)

					go triggerChainlinkWebhook(event.ReqId, event.DataFingerprint)
				}
			}
		}
		lastBlock = current
	}
}

// ── 被动接收（兼容旧方式）────────────────────────────────

type EventPayload struct {
	ReqId       string `json:"reqId"`
	Fingerprint string `json:"fingerprint"`
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var payload EventPayload
	json.Unmarshal(body, &payload)
	fmt.Printf("\n[Listener] 收到手动推送事件\n -> ReqID: %s\n -> CID: %s\n",
		payload.ReqId, payload.Fingerprint)
	go triggerChainlinkWebhook(payload.ReqId, payload.Fingerprint)
	w.WriteHeader(http.StatusOK)
}

// ── 入口 ──────────────────────────────────────────────────

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[Trigger] 配置加载失败: %v", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  跨链事件总线启动")
	fmt.Printf("  被动端口: %s\n", Cfg.Ports.AutoTrigger)
	fmt.Printf("  主动轮询: FISCO RPC %s (每 2 秒)\n", FiscoRPCURL)
	fmt.Printf("  目标合约: %s\n", Cfg.Fisco.ContractAddr)
	fmt.Printf("  监听事件: CrossDomainRequested\n")
	fmt.Printf("  Topic0:   %s\n", CrossDomainRequestedTopic)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	go startPolling()

	http.HandleFunc("/event", eventHandler)
	log.Fatal(http.ListenAndServe(Cfg.Ports.AutoTrigger, nil))
}
