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
	// 把旧的 2352c9c6... 换成截图里真实的 2d5857b0...
        ChainlinkWebhookURL = "http://localhost:6688/v2/jobs/2d5857b0-1f73-4a90-a33b-247e749c0c4d/runs" 
	OracleContractAddr  = "0xb59e0050aC3449D8A9F7A40670ed86DA7D89d5Ac"
	// ⚠️ 极其关键：请确保这里的账号密码，就是你登录 http://localhost:6688 的那个！
	CL_USER = "admin@crosschain.local"
	CL_PASS = "Admin@Chainlink2026"
)

func main() {
	fmt.Println("--------------------------------------------------")
	fmt.Println("[Listener] 正在初始化 FISCO BCOS 自动触发中枢 (Session 智能登录版)...")
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
			
			// ========================================================
			// ⚡ 核心测试区：强行喂入刚才在 Fabric 存好的真子弹 cert-888
			// ========================================================
			realCertID := "cert-888" 
			
			triggerChainlinkWebhook(realCertID)
		}
	}
}

// 接收动态 Hash 的触发函数 (自带智能登录破解版)
func triggerChainlinkWebhook(targetHash string) {
	fmt.Printf("[Trigger] 正在向 Chainlink 节点发送真实目标: %s ...\n", targetHash)

	// ==========================================
	// 🔑 动作 1: 先敲门登录，获取 Chainlink 的 Session 通行证
	// ==========================================
	loginReqBody := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, CL_USER, CL_PASS)
	loginResp, err := http.Post("http://localhost:6688/sessions", "application/json", bytes.NewBuffer([]byte(loginReqBody)))
	if err != nil || loginResp.StatusCode != 200 {
		fmt.Printf("[❌ 登录失败] 无法获取 Chainlink 授权 (请检查账号密码): %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	
	// 拿到极其珍贵的 Cookie (通行证)
	cookies := loginResp.Cookies() 

	// ==========================================
	// 🚀 动作 2: 带着通行证，把真子弹打进 Webhook
	// ==========================================
	requestBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{"hash": targetHash},
	})

	req, _ := http.NewRequest("POST", ChainlinkWebhookURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	// 把通行证塞进 HTTP 头里
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[❌ 轰炸失败] %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Printf("[✅ 轰炸成功] Chainlink 任务已激活! 状态码: %d\n\n", resp.StatusCode)
	} else {
		fmt.Printf("[⚠️ 异常] 状态码: %d (请检查 JobID 是否正确)\n", resp.StatusCode)
	}
}
