// listener/fisco_writer.go
// 使用 config.toml 配置
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type OracleResponse struct {
	ReqId        string `json:"reqId"`
	IsAuthorized bool   `json:"isAuthorized"`
	ResponseHash string `json:"responseHash"`
}

func executeConsoleCmd(method string, args ...string) (string, error) {
	cmdArgs := []string{"console.sh", "call",
		Cfg.Fisco.ContractName, Cfg.Fisco.ContractAddr, method}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, fmt.Sprintf("\"%s\"", arg))
	}
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = Cfg.Fisco.ConsoleDir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeBackHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var resp OracleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	authStr := "false"
	if resp.IsAuthorized {
		authStr = "true"
	}

	fmt.Printf("\n[Writer] 收到回写请求: ReqID=%s, Auth=%s, 时间戳: %d ms\n", resp.ReqId, authStr, time.Now().UnixMilli())
	out, err := executeConsoleCmd("fulfillAuth", resp.ReqId, authStr, resp.ResponseHash)
	if err != nil {
		fmt.Printf("[Writer] 回写失败: %v\n回执: %s\n", err, out)
		http.Error(w, out, http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Writer] 回写成功！数据闭环完成。时间戳: %d ms\n", time.Now().UnixMilli())
	w.WriteHeader(http.StatusOK)
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[Writer] 配置加载失败: %v", err)
	}
	fmt.Printf("[WriteBack] FISCO Writer 启动，监听 %s\n", Cfg.Ports.FiscoWriter)
	http.HandleFunc("/", writeBackHandler)
	log.Fatal(http.ListenAndServe(Cfg.Ports.FiscoWriter, nil))
}
