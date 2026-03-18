package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"listener/api"
)

const contractAddress = "0x77580befe1c74597dfdd626d6cced71807b868d4"
const adapterURL = "http://127.0.0.1:8081/"

type AdapterRequest struct {
	ID   string `json:"id"`
	Data struct {
		Hash string `json:"hash"`
	} `json:"data"`
}

type AdapterResponse struct {
	JobRunID string `json:"jobRunID"`
	Data     struct {
		IsValid bool `json:"isValid"`
	} `json:"data"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("[Fatal Error] 请提供交易哈希作为启动参数！")
		os.Exit(1)
	}
	txHash := os.Args[1]

	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		fmt.Printf("[Fatal Error] 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	c, err := client.Dial(&configs[0])
	if err != nil {
		fmt.Printf("[Fatal Error] FISCO连接失败: %v\n", err)
		os.Exit(1)
	}

	address := common.HexToAddress(contractAddress)
	filterer, err := api.NewCrossChainClientFilterer(address, nil)
	if err != nil {
		fmt.Printf("[Fatal Error] 解析器创建失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[Listener] 预言机大脑已切换为【主动截获模式】!")
	fmt.Printf("[Listener] 正在深度解析目标交易: %s\n", txHash)

	receipt, err := c.GetTransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		fmt.Printf("[Fatal Error] 无法获取交易回执: %v\n", err)
		os.Exit(1)
	}

	if len(receipt.Logs) == 0 {
		fmt.Println("[Fatal Error] 该交易未产生任何事件日志！")
		os.Exit(1)
	}

	for _, fiscoLog := range receipt.Logs {
		// 转换机制：将原始十六进制字符串翻译为标准字节流与哈希
		ethLog := ethTypes.Log{
			Data: common.FromHex(fiscoLog.Data),
		}
		for _, t := range fiscoLog.Topics {
			ethLog.Topics = append(ethLog.Topics, common.HexToHash(t))
		}

		event, parseErr := filterer.ParseCertVerificationRequested(ethLog)
		if parseErr != nil {
			continue
		}

		fmt.Printf("\n[Listener] 成功解析出跨链请求! ReqID: %x, Hash: %s\n", event.ReqId, event.CertHash)
		
		isValid := callFabricAdapter(event.ReqId, event.CertHash)

		fmt.Println("[Listener] 正在将权威结果写回 FISCO BCOS 账本...")

		reqIdHex := fmt.Sprintf("0x%x", event.ReqId)
		isValidStr := fmt.Sprintf("%t", isValid)
		cmd := exec.Command("bash", "/home/fan/console/console.sh", "call", "CrossChainClient", contractAddress, "fulfillVerification", reqIdHex, isValidStr)

		out, execErr := cmd.CombinedOutput()
		if execErr != nil {
			fmt.Printf("[Listener] 写入失败: %v\n回执: %s\n", execErr, string(out))
		} else {
			fmt.Printf("[Listener] 写入成功! 跨链验证结果已永久上链。\n%s\n", string(out))
		}
		return
	}
	
	fmt.Println("[Listener] 解析完毕，未在日志中匹配到跨链请求事件。")
}

func callFabricAdapter(reqId [32]byte, hash string) bool {
	reqBody := AdapterRequest{ID: fmt.Sprintf("%x", reqId)}
	reqBody.Data.Hash = hash
	jsonValue, _ := json.Marshal(reqBody)
	resp, reqErr := http.Post(adapterURL, "application/json", bytes.NewBuffer(jsonValue))
	if reqErr != nil {
		fmt.Printf("[Listener] 呼叫 Fabric 适配器失败: %v\n", reqErr)
		return false
	}
	defer resp.Body.Close()
	var adapterResp AdapterResponse
	json.NewDecoder(resp.Body).Decode(&adapterResp)
	fmt.Printf("[Listener] 收到 Fabric 权威裁决: %v\n", adapterResp.Data.IsValid)
	return adapterResp.Data.IsValid
}
