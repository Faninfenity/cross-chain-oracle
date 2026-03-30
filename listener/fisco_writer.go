package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

const ContractAddr = "0x0f3c7ca4308c17c6479d59c93d4095bb7032c781"
const ContractName = "CrossDomainAuth"

type OracleResponse struct {
	ReqId        string `json:"reqId"`
	IsAuthorized bool   `json:"isAuthorized"` 
	ResponseHash string `json:"responseHash"` 
}

func executeConsoleCmd(method string, args ...string) (string, error) {
	cmdArgs := []string{"console.sh", "call", ContractName, ContractAddr, method}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, fmt.Sprintf("\"%s\"", arg))
	}
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = "/home/fan/console"
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeBackHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var resp OracleResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("[Error] Parse payload failed: %v\n", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	authStr := "false"
	if resp.IsAuthorized {
		authStr = "true"
	}

	fmt.Printf("\n[Writer] Executing Phase 2: ReqID=%s, Auth=%s\n", resp.ReqId, authStr)

	out, err := executeConsoleCmd("fulfillAuth", resp.ReqId, authStr, resp.ResponseHash)
	
	if err != nil {
		fmt.Printf("[Writer Failed]: %v\nReceipt: %s\n", err, string(out))
		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Writer] Success! Data loop closed.\n")
	w.WriteHeader(http.StatusOK)
}

func main() {
	fmt.Println("[WriteBack] FISCO Writer started on port 8082")
	http.HandleFunc("/", writeBackHandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
