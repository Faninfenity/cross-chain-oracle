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

const FabricCliPath = "/home/fan/fabric-project/fabric-samples/test-network"

// 存证：将文件 CID 写入 Fabric 账本
func issueToFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少指纹参数", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("bash", "issue.sh", certHash)
	cmd.Dir = FabricCliPath
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err != nil {
		if strings.Contains(outStr, "already exists") {
			// 资产已存在，返回特殊标识让前端进一步判断
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "ALREADY_EXISTS")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Fabric 写入失败: %v\n回执: %s", err, outStr)
		}
		return
	}
	fmt.Fprintf(w, "成功录入 Fabric 账本: %s", outStr)
}

// 吊销：将指定 CID 的证书状态标记为 REVOKED
func revokeFromFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少指纹参数", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("bash", "revoke.sh", certHash)
	cmd.Dir = FabricCliPath
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err != nil {
		if strings.Contains(outStr, "证书不存在") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "吊销失败：该证书不存在于 Fabric 账本，请确认 CID 是否正确。")
		} else if strings.Contains(outStr, "已处于吊销状态") {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "吊销失败：该证书已处于吊销状态，不可重复吊销。")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "吊销操作执行失败: %v\n详细输出: %s", err, outStr)
		}
		return
	}

	if strings.Contains(outStr, "证书已吊销") || strings.Contains(outStr, "Chaincode invoke successful") {
		fmt.Fprintf(w, "REVOKE_SUCCESS:%s", certHash)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "吊销结果未知，请检查账本: %s", outStr)
	}
}

// 查询单个证书在 Fabric 的当前状态（供前端判断 already exists 时用）
func checkStatusHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少参数", http.StatusBadRequest)
		return
	}

	const (
		fabricBase = "/home/fan/fabric-project/fabric-samples"
		peerBin    = fabricBase + "/bin/peer"
		cfgPath    = fabricBase + "/config"
		tlsCert    = fabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
		mspPath    = fabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
	)

	cmd := exec.Command(peerBin, "chaincode", "query",
		"-C", "mychannel", "-n", "pki",
		"-c", fmt.Sprintf(`{"Args":["ReadAsset","%s"]}`, certHash))

	cmd.Env = []string{
		"PATH=" + fabricBase + "/bin:/usr/bin:/bin",
		"FABRIC_CFG_PATH=" + cfgPath,
		"CORE_PEER_TLS_ENABLED=true",
		"CORE_PEER_LOCALMSPID=Org1MSP",
		"CORE_PEER_TLS_ROOTCERT_FILE=" + tlsCert,
		"CORE_PEER_MSPCONFIGPATH=" + mspPath,
		"CORE_PEER_ADDRESS=localhost:7051",
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "NOTFOUND")
		return
	}
	fmt.Fprintf(w, "%s", strings.TrimSpace(string(out)))
}

// IPFS 上传：提取文件内容指纹 CID
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
        .drop-zone { width: 100%; height: 150px; border: 3px dashed #6366f1; border-radius: 12px; background: #1e1b4b; display: flex; flex-direction: column; align-items: center; justify-content: center; cursor: pointer; margin-bottom: 25px; box-sizing: border-box; transition: border-color 0.2s; }
        .drop-zone:hover { border-color: #a5b4fc; }
        .drop-zone-text { font-size: 18px; color: #e2e8f0; font-weight: bold; margin-bottom: 10px; }
        .hash-display-area { background: #0f172a; color: #facc15; padding: 15px; border-radius: 8px; border: 1px solid #334155; margin-bottom: 15px; font-family: 'Courier New', monospace; text-align: left; position: relative; }
        .hash-label { font-size: 12px; color: #94a3b8; position: absolute; top: -8px; left: 10px; background: #0f172a; padding: 0 5px; }
        .hash-value { font-size: 16px; font-weight: bold; word-break: break-all; }
        .hash-hidden { display: none; }
        .btn-issue { width: 100%; border: none; padding: 15px 0; border-radius: 8px; cursor: pointer; font-size: 16px; font-weight: bold; background: #4f46e5; color: white; margin-bottom: 15px; transition: background 0.2s; }
        .btn-issue:hover:not(:disabled) { background: #4338ca; }
        .btn-issue:disabled { opacity: 0.5; cursor: not-allowed; }
        .revoke-section { border: 1px solid #dc2626; border-radius: 8px; padding: 20px; margin-bottom: 20px; background: rgba(220, 38, 38, 0.08); }
        .revoke-title { color: #fca5a5; font-size: 13px; font-weight: bold; text-align: left; margin-bottom: 12px; letter-spacing: 1px; }
        .revoke-input-row { display: flex; gap: 10px; align-items: center; }
        .revoke-input { flex: 1; background: #0f172a; border: 1px solid #dc2626; color: #fca5a5; padding: 10px 14px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; outline: none; }
        .revoke-input::placeholder { color: #6b7280; }
        .btn-revoke { border: none; padding: 10px 20px; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: bold; background: #dc2626; color: white; white-space: nowrap; transition: background 0.2s; }
        .btn-revoke:hover { background: #b91c1c; }
        #log { background: #0f172a; color: #34d399; padding: 20px; border-radius: 8px; text-align: left; height: 280px; overflow-y: auto; font-family: 'Courier New', Courier, monospace; font-size: 14px; line-height: 1.7; border: 1px solid #334155; }
        .log-success { color: #34d399; }
        .log-error { color: #f87171; }
        .log-warn { color: #fbbf24; }
        .log-info { color: #60a5fa; }
        .divider { border: none; border-top: 1px solid #3730a3; margin: 20px 0; }
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
        <button class="btn-issue" id="issueBtn" onclick="issueToFabric()" disabled>
            [+] 确认颁发: 将指纹永久写入 Fabric 底层账本
        </button>

        <hr class="divider">

        <div class="revoke-section">
            <div class="revoke-title">⚠ 证书吊销管理 — 操作不可逆，请谨慎执行</div>
            <div class="revoke-input-row">
                <input type="text" class="revoke-input" id="revokeInput"
                    placeholder="输入待吊销证书的 IPFS CID...">
                <button class="btn-revoke" onclick="revokeFromFabric()">执行吊销</button>
            </div>
        </div>

        <div id="log">
            <span class="log-info">[系统] 存证中心初始化完成，等待铸造指令...</span>
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

        function appendLog(msg, type) {
            const classMap = { success: 'log-success', error: 'log-error', warn: 'log-warn', info: 'log-info' };
            const cls = classMap[type] || 'log-success';
            logEl.innerHTML += '<br><span class="' + cls + '">[日志] ' + msg + '</span>';
            logEl.scrollTop = logEl.scrollHeight;
        }

        async function handleFileSelected(file) {
            if (!file) return;
            appendLog('正在将源文件上传至 IPFS 星际文件系统...', 'info');
            const formData = new FormData();
            formData.append('file', file);
            try {
                const response = await fetch('/api/upload', { method: 'POST', body: formData });
                const result = await response.json();
                currentFileCID = result.Hash;
                currentHashSpan.textContent = currentFileCID;
                hashArea.classList.remove('hash-hidden');
                document.getElementById('revokeInput').value = currentFileCID;
                appendLog('IPFS 落块成功。CID: ' + currentFileCID, 'success');
                issueBtn.disabled = false;
            } catch (e) {
                appendLog('[错误] IPFS 上链失败: ' + e.message, 'error');
            }
        }

        dropZone.addEventListener('dragover', (e) => { e.preventDefault(); });
        dropZone.addEventListener('drop', (e) => { e.preventDefault(); handleFileSelected(e.dataTransfer.files[0]); });
        dropZone.addEventListener('click', () => { fileInput.click(); });
        fileInput.addEventListener('change', () => { handleFileSelected(fileInput.files[0]); });

        async function issueToFabric() {
            if (!currentFileCID) return;
            appendLog('正在执行原生脚本，向 Fabric 物理节点发送写事务...', 'info');
            try {
                const res = await fetch('/api/issue?id=' + currentFileCID);
                const text = await res.text();

                if (res.status === 200) {
                    appendLog('[成功] 账本固化完毕！该文件已获物理确权。', 'success');
                } else if (res.status === 409 && text === 'ALREADY_EXISTS') {
                    // 资产已存在，进一步查询是否已吊销
                    appendLog('[检查] 该 CID 已有存证记录，正在查询当前证书状态...', 'warn');
                    const checkRes = await fetch('/api/check?id=' + currentFileCID);
                    const checkText = await checkRes.text();

                    if (checkText.toUpperCase().includes('"COLOR":"REVOKED"') ||
                        checkText.toUpperCase().includes('"COLOR": "REVOKED"')) {
                        appendLog('[拒绝] 该证书已被吊销，不可重新颁发。', 'error');
                        appendLog('如需重新颁发，请修改文件内容生成新版本（产生新 CID）。', 'warn');
                    } else {
                        appendLog('[提示] 该证书已存在于账本且状态正常，无需重复颁发。', 'warn');
                    }
                } else {
                    appendLog('[错误] Fabric 录入回执: ' + text, 'error');
                }
            } catch (e) {
                appendLog('[错误] 网络连接异常: ' + e.message, 'error');
            }
        }

        async function revokeFromFabric() {
            const certHash = document.getElementById('revokeInput').value.trim();
            if (!certHash) {
                appendLog('[拦截] 请输入待吊销证书的 CID', 'warn');
                return;
            }
            if (!certHash.startsWith('Qm') && !certHash.startsWith('baf')) {
                appendLog('[拦截] CID 格式不合法，请检查输入', 'warn');
                return;
            }

            const confirmed = confirm(
                '⚠ 警告：证书吊销操作不可逆！\n\n' +
                '目标 CID: ' + certHash + '\n\n' +
                '吊销后该证书将在所有跨链查验中返回"已吊销"状态，\n' +
                '且该 CID 不可被重新颁发。\n\n' +
                '确认执行吊销操作？'
            );
            if (!confirmed) {
                appendLog('吊销操作已取消', 'warn');
                return;
            }

            appendLog('正在向 Fabric 节点发起吊销指令...', 'warn');
            try {
                const res = await fetch('/api/revoke?id=' + certHash);
                const text = await res.text();

                if (res.status === 200 && text.startsWith('REVOKE_SUCCESS')) {
                    appendLog('[成功] 证书已成功吊销！CID: ' + certHash, 'success');
                    appendLog('该证书在后续所有跨链核验中将返回"已吊销"判决。', 'warn');
                    appendLog('注意：该 CID 已永久锁定为吊销状态，不可重新颁发。', 'error');
                } else if (res.status === 404) {
                    appendLog('[失败] 证书不存在，请确认 CID 是否已存证', 'error');
                } else if (res.status === 409) {
                    appendLog('[失败] 该证书已处于吊销状态，不可重复吊销', 'error');
                } else {
                    appendLog('[失败] 吊销回执: ' + text, 'error');
                }
            } catch (e) {
                appendLog('[错误] 网络连接异常: ' + e.message, 'error');
            }
        }
    </script>
</body>
</html>`

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/upload", uploadToIPFSHandler)
	http.HandleFunc("/api/issue", issueToFabricHandler)
	http.HandleFunc("/api/revoke", revokeFromFabricHandler)
	http.HandleFunc("/api/check", checkStatusHandler)
	fmt.Println("源头存证系统已启动，端口 8889")
	log.Fatal(http.ListenAndServe(":8889", nil))
}
