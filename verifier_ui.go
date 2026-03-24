package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"
)

const ConsolePath = "/home/fan/console"
const ContractAddr = "0xe8dfaab25d58ae0c41e16cb679737ac3c8f5dc05"

func executeConsoleCmd(method, arg string) (string, error) {
	// 组装命令
	cmd := exec.Command("bash", "console.sh", "call", "CrossChainClient", ContractAddr, method, fmt.Sprintf("\"%s\"", arg))
	// 强制重定向到真正的控制台目录
	cmd.Dir = "/home/fan/console"
	
	// 只执行一次！并把最干净的结果直接返回给前端
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func uploadToIPFSHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "无法读取文件", http.StatusBadRequest)
		return
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", handler.Filename)
	io.Copy(part, file)
	writer.Close()
	req, _ := http.NewRequest("POST", "http://localhost:5001/api/v0/add", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "IPFS 连接失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	out, _ := executeConsoleCmd("requestVerification", certHash)
	fmt.Fprintf(w, "%s", out)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	out, _ := executeConsoleCmd("getResult", certHash)
	result := "[等待中] 链上状态尚未更新，请等待跨链回写..."
	if strings.Contains(out, "true") {
		result = "[核验通过] 跨链确权成功：底层权威账本中存在此记录。"
	} else if strings.Contains(out, "false") {
		result = "[核验驳回] 跨链确权失败：文件遭篡改或查无此证。"
	}
	fmt.Fprintf(w, "%s", result)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlTemplate))
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>第三方跨链查证总线 (Verifier)</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, sans-serif; background-color: #0f172a; color: #e2e8f0; display: flex; flex-direction: column; align-items: center; padding-top: 50px; }
        .card { background: #1e293b; padding: 40px; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); width: 800px; text-align: center; border: 1px solid #334155; }
        h1 { color: #38bdf8; margin-top: 0; font-size: 32px; }
        .subtitle { font-size: 14px; color: #94a3b8; margin-bottom: 30px; letter-spacing: 1px; }
        .drop-zone { width: 100%; height: 150px; border: 3px dashed #475569; border-radius: 12px; background: #0f172a; display: flex; flex-direction: column; align-items: center; justify-content: center; cursor: pointer; margin-bottom: 25px; box-sizing: border-box; }
        .drop-zone-text { font-size: 18px; color: #e2e8f0; font-weight: bold; margin-bottom: 10px; }
        .hash-display-area { background: #020617; color: #f59e0b; padding: 15px; border-radius: 8px; border: 1px solid #334155; margin-bottom: 25px; font-family: 'Courier New', monospace; text-align: left; position: relative;}
        .hash-label { font-size: 12px; color: #94a3b8; position: absolute; top: -8px; left: 10px; background: #020617; padding: 0 5px; }
        .hash-value { font-size: 16px; font-weight: bold; word-break: break-all;}
        .hash-hidden { display: none; }
        .btn-group { display: flex; justify-content: center; gap: 10px; margin-bottom: 25px; }
        button { flex: 1; border: none; padding: 12px 0; border-radius: 8px; cursor: pointer; font-size: 15px; font-weight: bold; color: white;}
        button[disabled] { opacity: 0.5; cursor: not-allowed; }
        .trigger-btn { background: #3b82f6; }
        .query-btn { background: #10b981; }
        #log { background: #020617; color: #10b981; padding: 20px; border-radius: 8px; text-align: left; height: 250px; overflow-y: auto; font-family: 'Courier New', Courier, monospace; font-size: 14px; line-height: 1.6; border: 1px solid #334155;}
    </style>
</head>
<body>
    <div class="card">
        <h1>[第三方] 跨链预言机查证总线</h1>
        <div class="subtitle">面向用人单位开放 | 触发 FISCO BCOS 跨链智能合约</div>
        <div id="dropZone" class="drop-zone">
            <span class="drop-zone-text">载入待核验的文件以提取 IPFS 标识</span>
            <input type="file" id="fileInput" class="hash-hidden">
        </div>
        <div id="hashArea" class="hash-display-area hash-hidden">
            <span class="hash-label">解析出的 IPFS 标识 (CID)</span>
            <span class="hash-value" id="currentHash"></span>
        </div>
        <div class="btn-group">
            <button class="trigger-btn" id="triggerBtn" onclick="trigger()" disabled>[>] 发起跨链核验任务</button>
            <button class="query-btn" id="queryBtn" onclick="query()" disabled>[?] 调取总线最终反馈</button>
        </div>
        <div id="log">
            <span style="color:#38bdf8">[系统] 查证总线已连接，异构网络穿透准备就绪。</span>
        </div>
    </div>
    <script>
        const logEl = document.getElementById('log');
        const dropZone = document.getElementById('dropZone');
        const fileInput = document.getElementById('fileInput');
        const hashArea = document.getElementById('hashArea');
        const currentHashSpan = document.getElementById('currentHash');
        const triggerBtn = document.getElementById('triggerBtn');
        const queryBtn = document.getElementById('queryBtn');
        let currentFileCID = null;

        function appendLog(msg) { logEl.innerHTML += '<br>[日志] ' + msg; logEl.scrollTop = logEl.scrollHeight; }

        async function handleFileSelected(file) {
            if (!file) return;
            appendLog('正在分析文件指纹，计算星际网络坐标...');
            const formData = new FormData(); formData.append('file', file);
            try {
                const response = await fetch('/api/upload', { method: 'POST', body: formData });
                const result = await response.json();
                currentFileCID = result.Hash;
                currentHashSpan.textContent = currentFileCID;
                hashArea.classList.remove('hash-hidden');
                appendLog('指纹提取完成。目标查证 CID: ' + currentFileCID);
                triggerBtn.disabled = false;
                queryBtn.disabled = false;
            } catch (e) { appendLog('[错误] 特征提取失败: ' + e.message); }
        }

        dropZone.addEventListener('dragover', (e) => { e.preventDefault(); });
        dropZone.addEventListener('drop', (e) => { e.preventDefault(); handleFileSelected(e.dataTransfer.files[0]); });
        dropZone.addEventListener('click', () => { fileInput.click(); });
        fileInput.addEventListener('change', () => { handleFileSelected(fileInput.files[0]); });

        async function trigger() {
            if(!currentFileCID) return;
            appendLog('正在向前端侧链 FISCO BCOS 抛出核验事件，唤醒预言机...');
            try {
                const res = await fetch('/api/trigger?id=' + currentFileCID);
                const text = await res.text();
                if(text.includes('transaction hash')) { appendLog('[系统] 事件抛出成功。跨链网关正在穿透中，请稍后查询最终状态。'); } 
                else { appendLog('[错误] 侧链调用失败: ' + text); }
            } catch (e) { appendLog('[错误] 网络连接异常'); }
        }

        async function query() {
            if(!currentFileCID) return;
            appendLog('正在通过控制台读取 FISCO BCOS 账本最新回写状态...');
            try {
                const res = await fetch('/api/query?id=' + currentFileCID);
                const text = await res.text();
                if(text.includes('核验通过')) appendLog('<span style="color:#10b981; font-weight:bold;">' + text + '</span>');
                else if(text.includes('核验驳回')) appendLog('<span style="color:#ef4444; font-weight:bold;">' + text + '</span>');
                else appendLog('<span style="color:#f59e0b;">' + text + '</span>');
            } catch (e) { appendLog('[错误] 网络连接异常'); }
        }
    </script>
</body>
</html>`

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/upload", uploadToIPFSHandler)
	http.HandleFunc("/api/trigger", triggerHandler)
	http.HandleFunc("/api/query", queryHandler)
	fmt.Println("跨链查证系统已启动，端口 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
