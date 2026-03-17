package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type NodeIdentity struct {
	OrgName string `json:"orgName"`
	Score   int    `json:"score"`
	IsActive bool  `json:"isActive"`
}

type Proposal struct {
	ProposalID   string            `json:"proposalID"`
	TargetNode   string            `json:"targetNode"`
	VotesFor     int               `json:"votesFor"`
	VotesAgainst int               `json:"votesAgainst"`
	Voters       map[string]bool   `json:"voters"`
	Status       string            `json:"status"`
}

func (s *SmartContract) InitSystem(ctx contractapi.TransactionContextInterface) error {
	nodes := []string{"Org1MSP", "Org2MSP", "Org3MSP", "Org4MSP"}
	for _, name := range nodes {
		node := NodeIdentity{OrgName: name, Score: 100, IsActive: true}
		nodeJSON, _ := json.Marshal(node)
		ctx.GetStub().PutState("NODE_"+name, nodeJSON)
	}
	return nil
}

func (s *SmartContract) CreateProposal(ctx contractapi.TransactionContextInterface, id string, target string) error {
	p := Proposal{ProposalID: id, TargetNode: target, Voters: make(map[string]bool), Status: "PENDING"}
	pJSON, _ := json.Marshal(p)
	return ctx.GetStub().PutState("PROP_"+id, pJSON)
}

func (s *SmartContract) Vote(ctx contractapi.TransactionContextInterface, id string, voter string, ok bool) error {
	pJSON, _ := ctx.GetStub().GetState("PROP_" + id)
	if pJSON == nil { return fmt.Errorf("提案不存在") }
	var p Proposal
	json.Unmarshal(pJSON, &p)
	if p.Voters[voter] { return fmt.Errorf("节点 %s 已投票", voter) }
	p.Voters[voter] = true
	if ok { p.VotesFor++ } else { p.VotesAgainst++ }
	if p.VotesFor >= 3 { p.Status = "APPROVED" }
	pJSON, _ = json.Marshal(p)
	return ctx.GetStub().PutState("PROP_"+id, pJSON)
}

// ======= 升级新增：查询单个提案 =======
func (s *SmartContract) QueryProposal(ctx contractapi.TransactionContextInterface, id string) (*Proposal, error) {
	pJSON, err := ctx.GetStub().GetState("PROP_" + id)
	if err != nil || pJSON == nil { return nil, fmt.Errorf("未找到提案 %s", id) }
	var p Proposal
	json.Unmarshal(pJSON, &p)
	return &p, nil
}

// ======= 升级新增：查询所有节点信誉 =======
func (s *SmartContract) QueryAllNodes(ctx contractapi.TransactionContextInterface) ([]*NodeIdentity, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("NODE_", "NODE_Z")
	if err != nil { return nil, err }
	defer resultsIterator.Close()
	var nodes []*NodeIdentity
	for resultsIterator.HasNext() {
		res, _ := resultsIterator.Next()
		var node NodeIdentity
		json.Unmarshal(res.Value, &node)
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

func main() {
	cc, _ := contractapi.NewChaincode(&SmartContract{})
	cc.Start()
}
