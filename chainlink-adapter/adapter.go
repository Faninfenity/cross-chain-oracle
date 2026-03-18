package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ChainlinkRequest struct {
	ID   string `json:"id"`
	Data struct {
		Hash string `json:"hash"`
	} `json:"data"`
}

type ChainlinkResponse struct {
	JobRunID string `json:"jobRunID"`
	Data     struct {
		IsValid bool `json:"isValid"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}

const (
	channelName   = "mychannel"
	chaincodeName = "realcert"
	mspID         = "Org1MSP"
	cryptoPath    = "/home/fan/fabric-project/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	certPath      = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath       = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath   = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint  = "127.0.0.1:7051"
	gatewayPeer   = "peer0.org1.example.com"
)

var contract *client.Contract

func main() {
	fmt.Println("[Adapter] 初始化 Fabric Gateway 长连接...")
	initFabricGateway()

	http.HandleFunc("/", handleRequest)
	fmt.Println("\n[Adapter] Chainlink 外部适配器已启动，监听端口 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func initFabricGateway() {
	clientConnection := newGrpcConnection()
	id := newIdentity()
	sign := newSign()

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(30 * time.Second),
		client.WithEndorseTimeout(30 * time.Second),
		client.WithSubmitTimeout(30 * time.Second),
		client.WithCommitStatusTimeout(1 * time.Minute),
	)
	if err != nil {
		panic(fmt.Errorf("failed to connect to gateway: %w", err))
	}

	network := gw.GetNetwork(channelName)
	contract = network.GetContract(chaincodeName)
	fmt.Println("[Adapter] Fabric Gateway 连接建立成功!")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var req ChainlinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("\n[Adapter] Received request from Chainlink! Job ID: %s, Hash: %s\n", req.ID, req.Data.Hash)

	isValid := false
	evaluateResult, err := contract.EvaluateTransaction("queryCertificate", req.Data.Hash)
	if err != nil {
		fmt.Printf("[Fabric Error] 查证失败或不存在: %v\n", err)
	} else {
		if len(evaluateResult) > 0 {
			isValid = true
			fmt.Printf("[Fabric Success] 确权成功! 返回内容: %s\n", string(evaluateResult))
		}
	}

	resp := ChainlinkResponse{
		JobRunID: req.ID,
	}
	resp.Data.IsValid = isValid

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	fmt.Println("[Adapter] Response sent back to Chainlink.")
}

func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)
	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}
	return connection
}

func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}
	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}
	return id
}

func newSign() identity.Sign {
	files, err := ioutil.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := ioutil.ReadFile(path.Join(keyPath, files[0].Name()))
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}
	return sign
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}
