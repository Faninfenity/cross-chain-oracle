package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PKIContract struct {
	contractapi.Contract
}

type Certificate struct {
	CertID    string `json:"certID"`
	Owner     string `json:"owner"`
	IPFSHash  string `json:"ipfsHash"` 
	Status    string `json:"status"`   
}

func (s *PKIContract) RequestCertVerification(ctx contractapi.TransactionContextInterface, certID string, owner string, ipfsHash string) error {
	cert := Certificate{
		CertID:   certID,
		Owner:    owner,
		IPFSHash: ipfsHash,
		Status:   "PENDING", 
	}
	certJSON, _ := json.Marshal(cert)
	err := ctx.GetStub().PutState(certID, certJSON)
	if err != nil { return err }

	eventPayload := []byte(fmt.Sprintf(`{"certID":"%s", "ipfsHash":"%s"}`, certID, ipfsHash))
	return ctx.GetStub().SetEvent("VerifyIPFSCert", eventPayload)
}

func (s *PKIContract) OracleCallback(ctx contractapi.TransactionContextInterface, certID string, isValid bool) error {
	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil || certJSON == nil { return fmt.Errorf("证书不存在") }

	var cert Certificate
	json.Unmarshal(certJSON, &cert)

	if isValid { cert.Status = "VERIFIED" } else { cert.Status = "REJECTED" }

	certJSON, _ = json.Marshal(cert)
	return ctx.GetStub().PutState(certID, certJSON)
}

func (s *PKIContract) QueryCertificate(ctx contractapi.TransactionContextInterface, certID string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil || certJSON == nil { return nil, fmt.Errorf("证书不存在") }
	var cert Certificate
	json.Unmarshal(certJSON, &cert)
	return &cert, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&PKIContract{})
	if err != nil { return }
	if err := chaincode.Start(); err != nil { return }
}
