package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PKIContract struct {
	contractapi.Contract
}

type Certificate struct {
	CertID    string `json:"certID"`
	Owner     string `json:"owner"`
	IPFSHash  string `json:"ipfsHash"`
	Status    string `json:"status"`    // VALID / REVOKED
	IssuedAt  string `json:"issuedAt"`
	RevokedAt string `json:"revokedAt"` // 空字符串表示未吊销
}

// CreateCert 存证：写入新证书，初始状态 VALID
func (s *PKIContract) CreateCert(ctx contractapi.TransactionContextInterface,
	certID string, owner string, ipfsHash string) error {

	// 防重复：已存在则拒绝
	existing, err := ctx.GetStub().GetState(certID)
	if err != nil {
		return fmt.Errorf("查询账本失败: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("ALREADY_EXISTS: 证书 %s 已存在", certID)
	}

	cert := Certificate{
		CertID:    certID,
		Owner:     owner,
		IPFSHash:  ipfsHash,
		Status:    "VALID",
		IssuedAt:  time.Now().UTC().Format(time.RFC3339),
		RevokedAt: "",
	}
	certJSON, _ := json.Marshal(cert)
	return ctx.GetStub().PutState(certID, certJSON)
}

// RevokeCert 吊销：将证书状态改为 REVOKED，不可逆
func (s *PKIContract) RevokeCert(ctx contractapi.TransactionContextInterface,
	certID string) error {

	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil {
		return fmt.Errorf("查询账本失败: %v", err)
	}
	if certJSON == nil {
		return fmt.Errorf("NOT_FOUND: 证书 %s 不存在", certID)
	}

	var cert Certificate
	json.Unmarshal(certJSON, &cert)

	// 不可逆约束：已吊销不可再次吊销
	if cert.Status == "REVOKED" {
		return fmt.Errorf("ALREADY_REVOKED: 证书 %s 已处于吊销状态", certID)
	}

	cert.Status    = "REVOKED"
	cert.RevokedAt = time.Now().UTC().Format(time.RFC3339)

	certJSON, _ = json.Marshal(cert)
	return ctx.GetStub().PutState(certID, certJSON)
}

// QueryCert 查询：返回完整证书记录
func (s *PKIContract) QueryCert(ctx contractapi.TransactionContextInterface,
	certID string) (*Certificate, error) {

	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil {
		return nil, fmt.Errorf("查询账本失败: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("NOT_FOUND: 证书 %s 不存在", certID)
	}

	var cert Certificate
	json.Unmarshal(certJSON, &cert)
	return &cert, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&PKIContract{})
	if err != nil {
		return
	}
	if err := chaincode.Start(); err != nil {
		return
	}
}
