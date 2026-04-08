// listener/auto_trigger.go
// 使用 config.toml 配置
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type EventPayload struct {
	ReqId       string `json:"reqId"`
	Fingerprint string `json:"fingerprint"`
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var payload EventPayload
	json.Unmarshal(body, &payload)

	fmt.Printf("\n[Listener] 事件捕获!\n -> ReqID: %s\n -> Fingerprint: %s\n",
		payload.ReqId, payload.Fingerprint)

	triggerChainlinkWebhook(payload.ReqId, payload.Fingerprint)
	w.WriteHeader(http.StatusOK)
}

func triggerChainlinkWebhook(reqId string, fingerprint string) {
	fmt.Println("[Oracle] 正在向 Chainlink 节点鉴权...")

	loginData := map[string]string{
		"email":    Cfg.Chainlink.Email,
		"password": Cfg.Chainlink.Password,
	}
	loginBytes, _ := json.Marshal(loginData)
	loginReq, _ := http.NewRequest("POST", Cfg.Chainlink.SessionURL(), bytes.NewBuffer(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	loginResp, err := client.Do(loginReq)
	if err != nil || loginResp.StatusCode != 200 {
		log.Printf("[Error] Chainlink 登录失败: %v\n", err)
		return
	}
	defer loginResp.Body.Close()

	var sessionCookie string
	for _, cookie := range loginResp.Cookies() {
		sessionCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	fmt.Println("[Oracle] 鉴权成功，Cookie 已获取")

	payloadData := map[string]string{
		"reqId":    reqId,
		"certHash": fingerprint,
	}
	payloadBytes, _ := json.Marshal(payloadData)

	req, _ := http.NewRequest("POST", Cfg.Chainlink.WebhookURL(), bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", sessionCookie)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Error] Webhook 触发失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[Oracle] Chainlink 请求已发送，HTTP Status: %d\n", resp.StatusCode)
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[Trigger] 配置加载失败: %v", err)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf("[Listener] 跨链事件总线启动，监听 %s\n", Cfg.Ports.AutoTrigger)
	fmt.Println("--------------------------------------------------")
	http.HandleFunc("/event", eventHandler)
	log.Fatal(http.ListenAndServe(Cfg.Ports.AutoTrigger, nil))
}
