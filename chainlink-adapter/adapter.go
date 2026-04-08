// chainlink-adapter/adapter.go
// 使用 config.toml 配置，不再有任何硬编码路径
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Certificate struct {
	CertID    string `json:"certID"`
	Status    string `json:"status"`
	IPFSHash  string `json:"ipfsHash"`
	Owner     string `json:"owner"`
	IssuedAt  string `json:"issuedAt"`
	RevokedAt string `json:"revokedAt"`
}

type CertStatus int

const (
	StatusValid    CertStatus = iota
	StatusRevoked
	StatusNotFound
)

func queryFabricLedger(targetHash string) (CertStatus, string) {
	args := fmt.Sprintf(`{"Args":["QueryCert", "%s"]}`, targetHash)
	cmd := exec.Command(Cfg.Fabric.PeerBin(), "chaincode", "query",
		"-C", Cfg.Fabric.Channel,
		"-n", Cfg.Fabric.Chaincode,
		"-c", args)

	cmd.Env = append(os.Environ(), Cfg.Fabric.FabricEnv()...)

	fmt.Printf("[Fabric] 穿透查询: %s\n", targetHash)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		fmt.Printf("[Fabric] 查询失败: %s\n", outputStr)
		return StatusNotFound, ""
	}

	fmt.Printf("[Fabric] 返回数据: %s\n", outputStr)

	var cert Certificate
	if err := json.Unmarshal([]byte(outputStr), &cert); err != nil {
		fmt.Printf("[Fabric] JSON解析失败: %v\n", err)
		return StatusNotFound, outputStr
	}

	switch cert.Status {
	case "VALID":
		return StatusValid, outputStr
	case "REVOKED":
		return StatusRevoked, outputStr
	default:
		return StatusNotFound, outputStr
	}
}

func handleChainlinkRequest(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	fmt.Printf("\n[X-Ray] 收到载荷: %s\n", string(bodyBytes))

	var req map[string]interface{}
	json.Unmarshal(bodyBytes, &req)
	hash, _ := req["certHash"].(string)
	reqId, _ := req["reqId"].(string)

	fmt.Printf("[Adapter] 目标指纹: %s\n", hash)
	status, _ := queryFabricLedger(hash)

	var isValid bool
	var responseHash, statusStr string
	switch status {
	case StatusValid:
		isValid, responseHash, statusStr = true, "Fabric-Status-VALID", "有效 (权威确权)"
	case StatusRevoked:
		isValid, responseHash, statusStr = false, "Fabric-Status-REVOKED", "已吊销"
	default:
		isValid, responseHash, statusStr = false, "Fabric-Status-NOTFOUND", "无效 (未注册或伪造)"
	}

	fmt.Printf("[Adapter] 最终判决: [%s]\n", statusStr)
	response := map[string]interface{}{
		"reqId":        reqId,
		"isAuthorized": isValid,
		"responseHash": responseHash,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	fmt.Println("[Adapter] 判决已回传 Chainlink 节点")
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[Adapter] 配置加载失败: %v", err)
	}
	http.HandleFunc("/", handleChainlinkRequest)
	fmt.Printf("[Adapter] 启动，监听 %s\n", Cfg.Ports.FabricAdapter)
	log.Fatal(http.ListenAndServe(Cfg.Ports.FabricAdapter, nil))
}
