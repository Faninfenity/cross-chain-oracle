package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
)

const (
	const ChainlinkWebhookURL = "http://localhost:6688/v2/jobs/2352c9c6-0e1c-4b2e-b867-c86cbcb96820/runs" 
	OracleContractAddr  = "0xb59e0050aC3449D8A9F7A40670ed86DA7D89d5Ac"
	// 认证信息
	CL_USER = "admin@crosschain.local"
	CL_PASS = "Admin@Chainlink2026"
)

func main() {
	fmt.Println("--------------------------------------------------")
	fmt.Println("[Listener] 正在初始化 FISCO BCOS 自动触发中枢 (带认证版)...")
	fmt.Println("--------------------------------------------------")

	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil { log.Fatalf("[错误] 解析配置失败: %v", err) }

	c, err := client.Dial(&configs[0])
	if err != nil { log.Fatalf("[错误] 无法连接 FISCO: %v", err) }
	defer c.Close()
	fmt.Println("[OK] FISCO 链路通畅，开始监听...")

	lastBlockNumber := getLatestBlockNumber(c)
	for {
		currentBlockNumber := getLatestBlockNumber(c)
		if currentBlockNumber > lastBlockNumber {
			fmt.Printf("[区块更新] 发现新高度: %d\n", currentBlockNumber)
			scanBlockForEvents(c, currentBlockNumber)
			lastBlockNumber = currentBlockNumber
		}
		time.Sleep(1 * time.Second)
	}
}

func getLatestBlockNumber(c *client.Client) int64 {
	bn, _ := c.GetBlockNumber(context.Background())
	return bn
}

func scanBlockForEvents(c *client.Client, blockNumber int64) {
	block, _ := c.GetBlockByNumber(context.Background(), blockNumber, true)
	if block == nil { return }

	for _, txInterface := range block.Transactions {
		tx, ok := txInterface.(map[string]interface{})
		if !ok { continue }
		txHash, _ := tx["hash"].(string)
		receipt, _ := c.GetTransactionReceipt(context.Background(), common.HexToHash(txHash))
		
		if receipt != nil && strings.EqualFold(receipt.To, OracleContractAddr) && receipt.Status == 0 {
			fmt.Printf("\n[🚨 拦截成功] 发现来自合约的跨链请求! TX: %s\n", txHash)
			triggerChainlinkWebhook()
		}
	}
}

func triggerChainlinkWebhook() {
	fmt.Println("[Trigger] 正在向 Chainlink 节点发起认证请求...")
	requestBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{"hash": "QmAutoTriggerHash999"},
	})

	// 创建带认证的请求
	req, _ := http.NewRequest("POST", ChainlinkWebhookURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	// 注入 Basic Auth 准入证
	req.SetBasicAuth(CL_USER, CL_PASS)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[❌ 失败] %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Printf("[✅ 轰炸成功] Chainlink 任务已激活! 状态码: %d\n\n", resp.StatusCode)
	} else {
		fmt.Printf("[⚠️ 依然异常] 状态码: %d (请检查 JobID 是否正确)\n", resp.StatusCode)
	}
}
