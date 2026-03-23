package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ChainlinkWebhookURL = "http://localhost:6688/v2/jobs/2d5857b0-1f73-4a90-a33b-247e749c0c4d/runs" 
	OracleContractAddr  = "0xe8dfaab25d58ae0c41e16cb679737ac3c8f5dc05"
	CL_USER = "admin@crosschain.local"
	CL_PASS = "Admin@Chainlink2026"
)

func main() {
	fmt.Println("--------------------------------------------------")
	fmt.Println("[Listener] 正在初始化 FISCO BCOS 自动触发中枢 (🚀 动态参数解析版)...")
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

// 💥 核心黑科技：底层 ABI 十六进制解码器！
// 专门用来把控制台传进来的 0x.... 乱码剥离回明文单号！
func decodeStringFromTxInput(hexData string) string {
	hexData = strings.TrimPrefix(hexData, "0x")
	// 剔除前 8 个字符（这是函数签名 Method ID）
	if len(hexData) < 8 { return "" }
	dataStr := hexData[8:]
	
	// 跳过 64 个字符的偏移量
	if len(dataStr) < 64 { return "" }
	// 提取 64 个字符的长度位
	if len(dataStr) < 128 { return "" }
	lengthHex := dataStr[64:128]
	
	// 计算字符串真实长度
	length, err := strconv.ParseInt(lengthHex, 16, 64)
	if err != nil || length == 0 { return "" }
	
	// 根据长度截取真实的 16 进制字符串并转码
	strDataHex := dataStr[128:]
	if int64(len(strDataHex)) < length*2 { return "" }
	actualStringHex := strDataHex[:length*2]
	
	bytesData, err := hex.DecodeString(actualStringHex)
	if err != nil { return "" }
	
	return string(bytesData)
}

func scanBlockForEvents(c *client.Client, blockNumber int64) {
	block, _ := c.GetBlockByNumber(context.Background(), blockNumber, true)
	if block == nil { return }

	for _, txInterface := range block.Transactions {
		tx, ok := txInterface.(map[string]interface{})
		if !ok { continue }
		
		txHash, _ := tx["hash"].(string)
		txInput, _ := tx["input"].(string) 

		receipt, _ := c.GetTransactionReceipt(context.Background(), common.HexToHash(txHash))
		
		if receipt != nil && strings.EqualFold(receipt.To, OracleContractAddr) && receipt.Status == 0 {
			
			// 防死循环装甲：忽略超长的回写交易
			if len(txInput) > 230 {
				continue
			}

			fmt.Printf("\n[🚨 拦截成功] 发现跨链请求! TX: %s\n", txHash)
			
			// ========================================================
			// 🎯 战役一通关：彻底废弃写死的 cert-888，启用动态解析！
			// ========================================================
			realCertID := decodeStringFromTxInput(txInput)
			if realCertID == "" {
				fmt.Println("[⚠️ 警告] 无法从交易底层解析出有效单号，已跳过。")
				continue
			}

			fmt.Printf("[🎯 动态锁定] 成功剥离底层 Hex 数据，目标单号: [%s]\n", realCertID)
			triggerChainlinkWebhook(realCertID)
		}
	}
}

func triggerChainlinkWebhook(targetHash string) {
	fmt.Printf("[Trigger] 正在向 Chainlink 节点发送真实目标: %s ...\n", targetHash)

	loginReqBody := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, CL_USER, CL_PASS)
	loginResp, err := http.Post("http://localhost:6688/sessions", "application/json", bytes.NewBuffer([]byte(loginReqBody)))
	if err != nil || loginResp.StatusCode != 200 {
		fmt.Printf("[❌ 登录失败] 无法获取授权: %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	
	cookies := loginResp.Cookies() 

	requestBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{"hash": targetHash},
	})

	req, _ := http.NewRequest("POST", ChainlinkWebhookURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies { req.AddCookie(cookie) }

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[❌ 轰炸失败] %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Printf("[✅ 轰炸成功] Chainlink 任务已激活! 状态码: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("[⚠️ 异常] 状态码: %d\n", resp.StatusCode)
	}
}
