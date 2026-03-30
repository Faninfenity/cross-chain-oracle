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

// 已经替换为你真实的 External Job ID
const WebhookURL = "http://localhost:6688/v2/jobs/2d5857b0-1f73-4a90-a33b-247e749c0c4d/runs"

type EventPayload struct {
	ReqId       string `json:"reqId"`
	Fingerprint string `json:"fingerprint"`
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var payload EventPayload
	json.Unmarshal(body, &payload)

	fmt.Printf("\n[Listener] Application-level event intercepted!\n")
	fmt.Printf(" -> ReqID: %s\n", payload.ReqId)
	fmt.Printf(" -> Fingerprint: %s\n", payload.Fingerprint)

	triggerChainlinkWebhook(payload.ReqId, payload.Fingerprint)
	w.WriteHeader(http.StatusOK)
}

func triggerChainlinkWebhook(reqId string, fingerprint string) {
	fmt.Printf("[Oracle] Authenticating with REAL Chainlink Node...\n")

	// 1. 自动登录 Chainlink 获取鉴权 Cookie
	loginURL := "http://localhost:6688/sessions"
	loginData := map[string]string{
		"email":    "admin@crosschain.local",
		"password": "Admin@Chainlink2026",
	}
	loginBytes, _ := json.Marshal(loginData)

	loginReq, _ := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBytes))
	loginReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Printf("[Error] Failed to connect to Chainlink: %v\n", err)
		return
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != 200 {
		log.Printf("[Error] Chainlink login failed, status: %d\n", loginResp.StatusCode)
		return
	}

	var sessionCookie string
	for _, cookie := range loginResp.Cookies() {
		sessionCookie += cookie.Name + "=" + cookie.Value + ";"
	}
	fmt.Printf("[Oracle] Auth successful! Session Cookie acquired.\n")

	// 2. 携带令牌，真正唤醒 Webhook 跨链任务
	payloadData := map[string]string{
		"reqId":    reqId,
		"certHash": fingerprint,
	}
	payloadBytes, _ := json.Marshal(payloadData)

	req, _ := http.NewRequest("POST", WebhookURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	// 注入登录凭证
	req.Header.Set("Cookie", sessionCookie)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Error] Failed to trigger Webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[Oracle] Real Chainlink request sent! HTTP Status: %d\n", resp.StatusCode)
}

func main() {
	fmt.Println("--------------------------------------------------")
	fmt.Println("[Listener] True Cross-Chain Event Bus started on :8083")
	fmt.Println("--------------------------------------------------")
	http.HandleFunc("/event", eventHandler)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
