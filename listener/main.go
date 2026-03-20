package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	fiscoAbi "github.com/FISCO-BCOS/go-sdk/abi"
	fiscoBind "github.com/FISCO-BCOS/go-sdk/abi/bind"
	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"listener/api"
)

const contractAddressHex = "0x77580befe1c74597dfdd626d6cced71807b868d4"
const adapterURL = "http://127.0.0.1:8081/"

// 极简版接口图纸，只包含我们需要打击的目标方法
const minABI = `[{"constant":false,"inputs":[{"name":"_reqId","type":"bytes32"},{"name":"_isValid","type":"bool"}],"name":"fulfillVerification","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

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
		log.Fatalf("[Fatal Error] 请提供交易哈希作为启动参数！\n")
	}
	txHashHex := os.Args[1]

	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		log.Fatalf("[Fatal Error] 配置加载失败: %v\n", err)
	}

	c, err := client.Dial(&configs[0])
	if err != nil {
		log.Fatalf("[Fatal Error] FISCO连接失败: %v\n", err)
	}

	contractAddress := common.HexToAddress(contractAddressHex)
	filterer, err := api.NewCrossChainClientFilterer(contractAddress, nil)
	if err != nil {
		log.Fatalf("[Fatal Error] 解析器创建失败: %v\n", err)
	}

	fmt.Println("[Listener] 预言机大脑已切换为主动截获模式")
	fmt.Printf("[Listener] 正在深度解析目标交易: %s\n", txHashHex)

	txHash := common.HexToHash(txHashHex)
	receipt, err := c.GetTransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatalf("[Fatal Error] 无法获取交易回执: %v\n", err)
	}

	if len(receipt.Logs) == 0 {
		log.Fatalf("[Fatal Error] 该交易未产生任何事件日志！\n")
	}

	for _, fiscoLog := range receipt.Logs {
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

		reqIdHex := fmt.Sprintf("%x", event.ReqId)
		fmt.Printf("\n[Listener] 成功解析出跨链请求! ReqID: 0x%s, Hash: %s\n", reqIdHex, event.CertHash)

		isValid := callFabricAdapter(event.ReqId, event.CertHash)

		fmt.Println("[Listener] 正在使用 FISCO 原生动态绑定引擎执行跨链存证回写...")

		// 1. 动态加载 ABI，完全抛弃静态生成的以太坊绑定代码
		parsedABI, err := fiscoAbi.JSON(strings.NewReader(minABI))
		if err != nil {
			log.Fatalf("[Fatal Error] ABI 解析失败: %v\n", err)
		}

		// 2. 利用 FISCO 客户端自身的能力，动态绑定目标合约
		boundContract := fiscoBind.NewBoundContract(contractAddress, parsedABI, c, c, c)
		auth := c.GetTransactOpts()

		// 3. 终极开火：无需拼接字符串，无需构造 Transaction 结构体，底层自动封包签名！
		_, txReceipt, err := boundContract.Transact(auth, "fulfillVerification", event.ReqId, isValid)
		if err != nil {
			log.Fatalf("[Listener] [Error] 原生回写上链失败: %v\n", err)
		}

		fmt.Printf("[Listener] [Success] 写入成功! 跨链验证结果已永久上链。\n")
		fmt.Printf("---------------- 回写存证凭证 ----------------\n")
		if txReceipt != nil {
			fmt.Printf("交易哈希: %s\n", txReceipt.TransactionHash)
			fmt.Printf("所在区块: %v\n", txReceipt.BlockNumber)
			fmt.Printf("状态码: %v (0x0 表示成功)\n", txReceipt.Status)
		}
		fmt.Printf("--------------------------------------------\n")

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
