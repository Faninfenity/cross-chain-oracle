package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
)

// 指向你刚写好的脚本目录
const FabricCliPath = "/home/fan/fabric-project/fabric-samples/test-network"

func issueToFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少指纹参数", http.StatusBadRequest)
		return
	}

	// 极其优雅：直接在宿主机调用刚写好的 bash 脚本，把哈希传进去
	cmd := exec.Command("bash", "issue.sh", certHash)
	cmd.Dir = FabricCliPath
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "Fabric 写入失败: %v\n回执: %s", err, string(out))
		return
	}
	fmt.Fprintf(w, "成功录入 Fabric 账本: %s", string(out))
}

func uploadToIPFSHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "无法读取上传的文件", http.StatusBadRequest)
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
		http.Error(w, "IPFS 节点连接失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlTemplate))
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>源头权威颁发中心 (Issuer)</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, sans-serif; background-color: #1e1b4b; color: #e2e8f0; display: flex; flex-direction: column; align-items: center; padding-top: 50px; }
        .card { background: #312e81; padding: 40px; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); width: 800px; text-align: center; border: 1px solid #4338ca; }
        h1 { color: #a5b4fc; margin-top: 0; font-size: 32px; }
        .subtitle { font-size: 14px; color: #c7d2fe; margin-bottom: 30px; letter-spacing: 1px; }
        .drop-zone { width: 100%; height: 150px; border: 3px dashed #6366f1; border-radius: 12px; background: #1e1b4b; display: flex; flex-direction: column; align-items: center; justify-content: center; cursor: pointer; margin-bottom: 25px; box-sizing: border-box; }
        .drop-zone-text { font-size: 18px; color: #e2e8f0; font-weight: bold; margin-bottom: 10px; }
        .hash-display-area { background: #0f172a; color: #facc15; padding: 15px; border-radius: 8px; border: 1px solid #334155; margin-bottom: 25px; font-family: 'Courier New', monospace; text-align: left; position: relative;}
        .hash-label { font-size: 12px; color: #94a3b8; position: absolute; top: -8px; left: 10px; background: #0f172a; padding: 0 5px; }
        .hash-value { font-size: 16px; font-weight: bold; word-break: break-all;}
        .hash-hidden { display: none; }
        button { width: 100%; border: none; padding: 15px 0; border-radius: 8px; cursor: pointer; font-size: 16px; font-weight: bold; background: #4f46e5; color: white; margin-bottom: 25px;}
        button[disabled] { opacity: 0.5; cursor: not-allowed; }
        #log { background: #0f172a; color: #34d399; padding: 20px; border-radius: 8px; text-align: left; height: 250px; overflow-y: auto; font-family: 'Courier New', Courier, monospace; font-size: 14px; line-height: 1.6; border: 1px solid #334155;}
    </style>
</head>
<body>
    <div class="card">
        <h1>[机构端] 源头权威存证系统</h1>
        <div class="subtitle">仅限授权机构访问 | 目标网络: IPFS + Hyperledger Fabric</div>
        <div id="dropZone" class="drop-zone">
            <span class="drop-zone-text">将待颁发的源文件拖拽至此进行哈希锁定</span>
            <input type="file" id="fileInput" class="hash-hidden">
        </div>
        <div id="hashArea" class="hash-display-area hash-hidden">
            <span class="hash-label">IPFS 全网唯一标识 (CID)</span>
            <span class="hash-value" id="currentHash"></span>
        </div>
        <button id="issueBtn" onclick="issueToFabric()" disabled>[+] 确认颁发: 将指纹永久写入 Fabric 底层账本</button>
        <div id="log">
            <span style="color:#60a5fa">[系统] 存证中心初始化完成，等待铸造指令...</span>
        </div>
    </div>
    <script>
        const logEl = document.getElementById('log');
        const dropZone = document.getElementById('dropZone');
        const fileInput = document.getElementById('fileInput');
        const hashArea = document.getElementById('hashArea');
        const currentHashSpan = document.getElementById('currentHash');
        const issueBtn = document.getElementById('issueBtn');
        let currentFileCID = null;

        function appendLog(msg) { logEl.innerHTML += '<br>[日志] ' + msg; logEl.scrollTop = logEl.scrollHeight; }

        async function handleFileSelected(file) {
            if (!file) return;
            appendLog('正在将源文件上传至 IPFS 星际文件系统...');
            const formData = new FormData(); formData.append('file', file);
            try {
                const response = await fetch('/api/upload', { method: 'POST', body: formData });
                const result = await response.json();
                currentFileCID = result.Hash;
                currentHashSpan.textContent = currentFileCID;
                hashArea.classList.remove('hash-hidden');
                appendLog('IPFS 落块成功。获得 CID: ' + currentFileCID);
                issueBtn.disabled = false;
            } catch (e) { appendLog('[错误] IPFS 上链失败: ' + e.message); }
        }

        dropZone.addEventListener('dragover', (e) => { e.preventDefault(); });
        dropZone.addEventListener('drop', (e) => { e.preventDefault(); handleFileSelected(e.dataTransfer.files[0]); });
        dropZone.addEventListener('click', () => { fileInput.click(); });
        fileInput.addEventListener('change', () => { handleFileSelected(fileInput.files[0]); });

        async function issueToFabric() {
            if(!currentFileCID) return;
            appendLog('正在执行原生脚本，向 Fabric 物理节点发送写事务...');
            try {
                const res = await fetch('/api/issue?id=' + currentFileCID);
                const text = await res.text();
                if(text.includes('Chaincode invoke successful') || text.includes('status:200') || text.includes('成功录入')) {
                    appendLog('[成功] 账本固化完毕！该文件已获物理确权。');
                } else { appendLog('[错误] Fabric 录入回执: ' + text); }
            } catch (e) { appendLog('[错误] 网络连接异常'); }
        }
    </script>
</body>
</html>`

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/upload", uploadToIPFSHandler)
	http.HandleFunc("/api/issue", issueToFabricHandler)
	fmt.Println("源头存证系统已启动，端口 8889")
	log.Fatal(http.ListenAndServe(":8889", nil))
}
