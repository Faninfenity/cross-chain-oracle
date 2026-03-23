package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

const SecurityToken = "LaoFan-CrossChain-King-2026"
const ConsolePath = "/home/fan/console"

// 🚨 已经为你替换成刚刚部署的全新地址！！！
const ContractAddr = "0xe8dfaab25d58ae0c41e16cb679737ac3c8f5dc05" 

func executeFiscoTransaction(targetID string, isValid bool) {
	validStr := "false"
	if isValid { validStr = "true" }

	// 🛡️ 给参数加上字面量的双引号，迎合 FISCO Java 控制台处理 string 的脾气
	quotedTargetID := fmt.Sprintf("\"%s\"", targetID)

	fmt.Printf("[WriteBack] ⚡ 正在唤醒 FISCO 底层控制台...\n")

	// 🛡️ 企业级核心：严格参数隔离！绝对禁止 bash -c 拼接！彻底封杀命令注入！
	cmd := exec.Command("bash", "console.sh", "call", "CrossChainClient", ContractAddr, "fulfillVerification", quotedTargetID, validStr)
	cmd.Dir = ConsolePath 

	output, err := cmd.CombinedOutput()
	outStr := string(output)

	// 把底层的真实回执打印出来，绝不吞没报错
	fmt.Printf("[WriteBack] 🔍 控制台底层回执:\n%s\n", strings.TrimSpace(outStr))

	if err != nil {
		fmt.Printf("[WriteBack] ❌ 唤醒控制台失败: %v\n", err)
		return
	}

	if strings.Contains(outStr, "transaction hash") {
		fmt.Printf("\n[WriteBack] ✅ 交易落块成功！物理状态已永久固化！\n")
		lines := strings.Split(outStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "transaction hash") {
				fmt.Printf("[WriteBack] 🧾 %s\n", strings.TrimSpace(line))
			}
		}
	} else {
		fmt.Printf("\n[WriteBack] ⚠️ 未发现交易哈希，可能是合约执行失败，请检查上面回执！\n")
	}
}

func handleWriteBack(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != SecurityToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	var targetID string
	var isValid bool

	if data, ok := req["data"].(map[string]interface{}); ok {
		if idStr, exists := data["id"].(string); exists { targetID = idStr }
		if valid, exists := data["isValid"].(bool); exists { isValid = valid }
	}

	fmt.Printf("\n==================================================\n")
	fmt.Printf("[WriteBack] 跨链目标: %s | Fabric 判决: %v\n", targetID, isValid)
	
	executeFiscoTransaction(targetID, isValid)

	fmt.Println("[WriteBack] 🎉 跨链全生命周期大闭环，彻底物理竣工！！！")
	fmt.Printf("==================================================\n\n")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func main() {
	http.HandleFunc("/", handleWriteBack)
	fmt.Println("[WriteBack] 🛡️ FISCO 物理回写中枢 (防注入+双引号对齐版) 已启动，监听 :8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
