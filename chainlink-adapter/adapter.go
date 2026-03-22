package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	FabricBase    = "/home/fan/fabric-project/fabric-samples"
	PeerBin       = FabricBase + "/bin/peer"
	CfgPath       = FabricBase + "/config"
	TlsCert       = FabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
	MspPath       = FabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
	PeerAddress   = "localhost:7051"

	ChannelName   = "mychannel"
	ChaincodeName = "pki"         
	QueryFuncName = "QueryCertificate"   // 真实的 Fabric 查询函数
	ListenPort    = ":8081"
)

func queryFabricLedger(targetHash string) (bool, error) {
	// 把它改成下面这样，把 QueryFuncName 加进去！
chaincodeArgs := fmt.Sprintf(`{"Args":["%s", "%s"]}`, QueryFuncName, targetHash)
	
	cmd := exec.Command(PeerBin, "chaincode", "query",
		"-C", ChannelName,
		"-n", ChaincodeName,
		"-c", chaincodeArgs)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FABRIC_CFG_PATH="+CfgPath)
	cmd.Env = append(cmd.Env, "CORE_PEER_TLS_ENABLED=true")
	cmd.Env = append(cmd.Env, "CORE_PEER_LOCALMSPID=Org1MSP")
	cmd.Env = append(cmd.Env, "CORE_PEER_TLS_ROOTCERT_FILE="+TlsCert)
	cmd.Env = append(cmd.Env, "CORE_PEER_MSPCONFIGPATH="+MspPath)
	cmd.Env = append(cmd.Env, "CORE_PEER_ADDRESS="+PeerAddress)

	fmt.Printf("[Fabric] 正在启动宿主机穿透查询...\n")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	if err != nil {
		fmt.Printf("[Fabric] 查无此证或链码报错:\n%s\n", outputStr)
		return false, nil 
	}

	fmt.Printf("[Fabric] 💥 底层返回原始数据: %s\n", strings.TrimSpace(outputStr))
	if strings.Contains(strings.ToLower(outputStr), "error") {
		return false, nil
	}
	return true, nil
}

func handleChainlinkRequest(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)
	
	jobID, _ := req["id"].(string)
	data, _ := req["data"].(map[string]interface{})
	hash, _ := data["hash"].(string)

	fmt.Printf("\n[Adapter] 收到跨链核查任务! JobID: %s, 目标Hash: %s\n", jobID, hash)

	isValid, _ := queryFabricLedger(hash)
	statusStr := "无效 (非法伪造)"
	if isValid { statusStr = "有效 (权威确权)" }
	fmt.Printf("[Adapter] 最终判决: [%s]\n", statusStr)

	response := map[string]interface{}{
		"jobRunID": jobID,
		"data": map[string]interface{}{"isValid": isValid},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	fmt.Println("[Adapter] 判决结果已成功打回 Chainlink 节点!")
}

func main() {
	http.HandleFunc("/", handleChainlinkRequest)
	fmt.Printf("[Adapter] 宿主机原生穿透版适配器已启动，监听 %s...\n", ListenPort)
	log.Fatal(http.ListenAndServe(ListenPort, nil))
}
