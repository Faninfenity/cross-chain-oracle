package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Chainlink 节点发来的标准请求体
type ChainlinkRequest struct {
	ID   string `json:"id"`
	Data struct {
		Hash string `json:"hash"`
	} `json:"data"`
}

// 必须严格遵守的 Chainlink 节点响应体
type ChainlinkResponse struct {
	JobRunID   string      `json:"jobRunID"`
	Data       interface{} `json:"data"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"statusCode"`
}

func main() {
	http.HandleFunc("/", handleRequest)
	
	fmt.Println("[Adapter] Chainlink 专属外部适配器已启动，监听端口 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 1. 读取并解析 Chainlink 发来的数据
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var req ChainlinkRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, req.ID, "无法解析 Chainlink JSON 请求体: "+err.Error())
		return
	}

	fmt.Printf("\n[Adapter] 收到 Chainlink 跨链核查任务! JobID: %s, 目标Hash: %s\n", req.ID, req.Data.Hash)

	// =========================================================================
	// 2. 这里是调用 Fabric 的核心逻辑
	// 为了确保当前环境能 100% 跑通，我们先用一个模拟的 Fabric 查证逻辑。
	// 等 Chainlink 流水线全线贯通后，咱们再把真正的 Fabric SDK 查证代码嵌进这里。
	// =========================================================================
	
	isValid := mockFabricVerify(req.Data.Hash)
	
	if isValid {
		fmt.Printf("[Adapter] Fabric 底层账本研判结果: [有效 - 证书存在]\n")
	} else {
		fmt.Printf("[Adapter] Fabric 底层账本研判结果: [无效 - 查无此证]\n")
	}

	// 3. 按照 Chainlink 的死板要求，组装返回格式
	resp := ChainlinkResponse{
		JobRunID: req.ID, // 必须原样返回 JobID
		Data: map[string]interface{}{
			"isValid": isValid, // 核心验证结果
		},
		StatusCode: 200,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
	
	fmt.Printf("[Adapter] 已将权威判决打包完毕，正打回 Chainlink 节点...\n")
}

// 模拟向 Fabric 发起查询的函数
func mockFabricVerify(hash string) bool {
	// 模拟耗时网络请求
	// time.Sleep(1 * time.Second)
	
	// 如果哈希是咱们设定的测试哈希，就返回 true，否则返回 false
	if hash == "QmTestHash123456789" {
		return true
	}
	return false
}

func sendError(w http.ResponseWriter, jobID string, errMsg string) {
	resp := ChainlinkResponse{
		JobRunID:   jobID,
		Error:      errMsg,
		StatusCode: 500,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}
