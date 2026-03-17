package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CertPayload struct {
	ID        string `json:"id"`
	Issuer    string `json:"issuer"`
	IssueDate string `json:"issueDate"`
	IsValid   bool   `json:"isValid"`
}

var contract *client.Contract

func main() {
	fmt.Println("=====================================================")
	fmt.Println("🚀 异构双链跨链预言机 (微服务 API 监听版) 启动中...")
	fmt.Println("=====================================================")

	initFabricConnection()

	http.HandleFunc("/api/verify", handleCrossChainRequest)

	port := ":8080"
	fmt.Printf("✅ [状态] 预言机已进入守护模式！正在监听端口 %s ...\n", port)
	fmt.Println("👉 (你可以随时通过浏览器或终端发送 HTTP 请求来触发跨链)")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("❌ Web 服务器崩溃: %v", err)
	}
}

func handleCrossChainRequest(w http.ResponseWriter, r *http.Request) {
	targetHash := r.URL.Query().Get("hash")
	if targetHash == "" {
		http.Error(w, "缺少 hash 参数", http.StatusBadRequest)
		return
	}

	fmt.Printf("\n🔔 [触发器] 收到跨链 API 请求，正在验证哈希: %s\n", targetHash)

	evaluateResult, err := contract.EvaluateTransaction("QueryCertificate", targetHash)
	if err != nil {
		errMsg := fmt.Sprintf("❌ 权威链查无此证: %v", err)
		fmt.Println("   " + errMsg)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	fmt.Println("   ✨【Fabric 侧】查证成功！权威链已确权！")

	var certData CertPayload
	json.Unmarshal(evaluateResult, &certData)

	fmt.Printf("   🚀 正在唤醒底层引擎，将状态 [%v] 写入 FISCO BCOS...\n", certData.IsValid)

	fiscoCmd := fmt.Sprintf(`cd ~/console && echo -e "call CertOracle 0x307b09cdfae62103bb3d3c361855cfbbc1d92158 writeBackCertStatus \"%s\" %v\nquit" | bash start.sh 1`, certData.ID, certData.IsValid)
	cmd := exec.Command("bash", "-c", fiscoCmd)
	_, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("   ❌ 写入 FISCO 失败: %v\n", err)
		http.Error(w, "写入 FISCO 失败", http.StatusInternalServerError)
		return
	}

	fmt.Println("   🏆【完美闭环】交易已成功发往 FISCO BCOS！")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON := fmt.Sprintf(`{"status": "success", "fisco_receipt": "写入成功", "hash": "%s", "valid": %v}`, certData.ID, certData.IsValid)
	w.Write([]byte(responseJSON))
}

func initFabricConnection() {
	mspDir := "/home/fan/fabric-project/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
	keyDir := filepath.Join(mspDir, "keystore")
	certDir := filepath.Join(mspDir, "signcerts")
	tlsCertPath := "/home/fan/fabric-project/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint := "localhost:7051"
	gatewayPeer := "peer0.org1.example.com"

	certFiles, _ := os.ReadDir(certDir)
	certPEM, _ := os.ReadFile(filepath.Join(certDir, certFiles[0].Name()))
	cert, _ := identity.CertificateFromPEM(certPEM)

	keyFiles, _ := os.ReadDir(keyDir)
	keyPEM, _ := os.ReadFile(filepath.Join(keyDir, keyFiles[0].Name()))
	key, _ := identity.PrivateKeyFromPEM(keyPEM)

	id, _ := identity.NewX509Identity("Org1MSP", cert)
	sign, _ := identity.NewPrivateKeySign(key)

	tlsCA, _ := os.ReadFile(tlsCertPath)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(tlsCA)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)
	grpcConn, _ := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))

	gw, _ := client.Connect(id, client.WithSign(sign), client.WithClientConnection(grpcConn))
	network := gw.GetNetwork("servicechannel")
	contract = network.GetContract("realcert")
}
