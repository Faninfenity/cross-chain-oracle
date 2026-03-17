package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
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

func main() {
	fmt.Println("=====================================================")
	fmt.Println("🚀 异构双链跨链预言机 (全自动大闭环完全体) 启动中...")
	fmt.Println("=====================================================")

	mspDir := "/home/fan/fabric-project/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
	keyDir := filepath.Join(mspDir, "keystore")
	certDir := filepath.Join(mspDir, "signcerts")
	tlsCertPath := "/home/fan/fabric-project/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint := "localhost:7051"
	gatewayPeer := "peer0.org1.example.com"

	certFiles, _ := os.ReadDir(certDir)
	certPath := filepath.Join(certDir, certFiles[0].Name())
	certPEM, _ := os.ReadFile(certPath)
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
	defer grpcConn.Close()

	gw, _ := client.Connect(id, client.WithSign(sign), client.WithClientConnection(grpcConn))
	defer gw.Close()

	network := gw.GetNetwork("servicechannel")
	contract := network.GetContract("realcert")

	fmt.Println("✅ [状态] 双链引擎点火完毕！进入全自动跨链搬运模式...")

	targetHash := "Qm1234567890abcdef"
	
	fmt.Printf("\n🔔 [跨链任务] 侦测到源链请求，正在跨链验证 IPFS 哈希: %s\n", targetHash)
	
	evaluateResult, err := contract.EvaluateTransaction("QueryCertificate", targetHash)
	if err != nil {
		fmt.Printf("   ❌ 权威链查无此证: %v\n", err)
		return
	} 
	
	fmt.Println("   ✨【Fabric 侧】查证成功！权威链已确权！")
	
	var certData CertPayload
	err = json.Unmarshal(evaluateResult, &certData)
	if err != nil {
		fmt.Printf("   ⚠️ 解析 JSON 失败: %v\n", err)
		return
	}

	fmt.Printf("   🚀 正在拉起控制台引擎，将合法状态 [%v] 写入 FISCO BCOS...\n", certData.IsValid)

	// 🌟 终极修正：用 echo -e 结合管道符，强行传入指令和 quit 退出信号，指定群组 1，完美避开所有参数坑！
	fiscoCmd := fmt.Sprintf(`cd ~/console && echo -e "call CertOracle 0x307b09cdfae62103bb3d3c361855cfbbc1d92158 writeBackCertStatus \"%s\" %v\nquit" | bash start.sh 1`, certData.ID, certData.IsValid)

	cmd := exec.Command("bash", "-c", fiscoCmd)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		fmt.Printf("   ❌ 写入 FISCO BCOS 失败: %v\n", err)
	} else {
		fmt.Println("   🏆【完美闭环】交易已成功发往 FISCO BCOS 并在前端链固化！")
		fmt.Printf("   👉 FISCO 底层回执简报:\n%s\n", string(output))
		fmt.Println("🎉🎉🎉 全系统打通！《异构双链跨链存证系统》MVP 运行圆满成功！")
	}
}
