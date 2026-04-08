package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const ContractAddr = "0x0f3c7ca4308c17c6479d59c93d4095bb7032c781"
const ContractName = "CrossDomainAuth"

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

func queryFabricDirect(certHash string) string {
	const (
		fabricBase = "/home/fan/fabric-project/fabric-samples"
		peerBin    = fabricBase + "/bin/peer"
		cfgPath    = fabricBase + "/config"
		tlsCert    = fabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
		mspPath    = fabricBase + "/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"
	)

	cmd := exec.Command(peerBin, "chaincode", "query",
		"-C", "mychannel",
		"-n", "pki",
		"-c", fmt.Sprintf(`{"Args":["QueryCert", "%s"]}`, certHash))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FABRIC_CFG_PATH="+cfgPath)
	cmd.Env = append(cmd.Env, "CORE_PEER_TLS_ENABLED=true")
	cmd.Env = append(cmd.Env, "CORE_PEER_LOCALMSPID=Org1MSP")
	cmd.Env = append(cmd.Env, "CORE_PEER_TLS_ROOTCERT_FILE="+tlsCert)
	cmd.Env = append(cmd.Env, "CORE_PEER_MSPCONFIGPATH="+mspPath)
	cmd.Env = append(cmd.Env, "CORE_PEER_ADDRESS=localhost:7051")

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if strings.Contains(outputStr, "NOT_FOUND") || strings.Contains(outputStr, "does not exist") {
			return "STATUS:NOTFOUND"
		}
		return "STATUS:ERROR:" + outputStr
	}

	// 解析 JSON 的 status 字段
	var cert struct {
		Status string `json:"status"`
	}
	if jsonErr := json.Unmarshal([]byte(outputStr), &cert); jsonErr != nil {
		return "STATUS:ERROR:JSON解析失败"
	}

	switch cert.Status {
	case "VALID":
		return "STATUS:VALID:" + outputStr
	case "REVOKED":
		return "STATUS:REVOKED:" + outputStr
	default:
		return "STATUS:NOTFOUND"
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("myFile")
	if err != nil {
		fmt.Fprintf(w, "[Error] File read failed: %v", err)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "cross-chain-*.tmp")
	if err != nil {
		fmt.Fprintf(w, "[Error] Temp file creation failed: %v", err)
		return
	}
	defer os.Remove(tempFile.Name())
	io.Copy(tempFile, file)
	tempFile.Close()

	cmd := exec.Command("ipfs", "add", "-Q", "-n", tempFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "[Error] IPFS CID computation failed: %v\n%s", err, string(out))
		return
	}

	fingerprint := strings.TrimSpace(string(out))
	reqId := fingerprint
	sourceDID := "did:5g:base-station-001"
	targetDomain := "core-network-domain"

	fmt.Printf("\n[UI] File uploaded, IPFS CID Extracted: %s\n", fingerprint)

	outCmd, _ := executeConsoleCmd("requestDataAuth", reqId, sourceDID, targetDomain, fingerprint)

	if strings.Contains(outCmd, "Request ID already exists") {
		fmt.Fprintf(w, "[智能路由] 拦截到重复请求！\n\nReqID: %s\n\n请直接点击【检索底层账本】查看历史权威判决。", reqId)
		return
	}

	fmt.Fprintf(w, "File CID extracted successfully!\nReqID: %s\n\nEvent Receipt:\n%s", reqId, outCmd)

	if strings.Contains(outCmd, "transaction executed successfully") {
		fmt.Printf("[UI] Chain write successful, pushing to event bus...\n")
		payload := fmt.Sprintf(`{"reqId":"%s", "fingerprint":"%s"}`, reqId, fingerprint)
		http.Post("http://localhost:8083/event", "application/json", strings.NewReader(payload))
	}
}

func queryStatusHandler(w http.ResponseWriter, r *http.Request) {
	reqId := r.URL.Query().Get("id")
	out, _ := executeConsoleCmd("getAuthStatus", reqId)
	fmt.Fprintf(w, "%s", out)
}

func queryFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少CID参数", http.StatusBadRequest)
		return
	}
	result := queryFabricDirect(certHash)
	fmt.Fprintf(w, "%s", result)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, html)
}

const html = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>5G-A 异构跨域信任协同矩阵</title>
    <style>
        :root { --primary: #00ffcc; --bg-dark: #0a0a0f; --panel-bg: rgba(16,20,30,0.85); --border-color: rgba(0,255,204,0.3); --text-main: #e0e0e0; --terminal-bg: #050505; }
        body { background-color: var(--bg-dark); background-image: linear-gradient(rgba(0,255,204,0.03) 1px,transparent 1px),linear-gradient(90deg,rgba(0,255,204,0.03) 1px,transparent 1px); background-size: 30px 30px; color: var(--text-main); font-family: 'Consolas','Courier New',monospace; margin: 0; padding: 40px; display: flex; flex-direction: column; align-items: center; }
        .header { text-align: center; margin-bottom: 40px; letter-spacing: 2px; color: var(--primary); border-bottom: 1px solid var(--primary); padding-bottom: 15px; width: 80%; max-width: 900px; }
        .container { width: 80%; max-width: 900px; display: grid; gap: 25px; }
        .panel { background: var(--panel-bg); border: 1px solid var(--border-color); border-radius: 4px; padding: 25px; position: relative; overflow: hidden; }
        .panel::before { content: ''; position: absolute; top: 0; left: 0; width: 4px; height: 100%; background: var(--primary); }
        h3 { margin-top: 0; color: #fff; font-size: 1.2em; letter-spacing: 1px; border-bottom: 1px dashed #333; padding-bottom: 10px; }
        .input-group { margin: 20px 0; display: flex; gap: 15px; align-items: center; }
        input[type="text"] { flex: 1; background: rgba(0,0,0,0.5); border: 1px solid var(--border-color); color: var(--primary); padding: 12px; font-family: inherit; outline: none; }
        .file-upload-wrapper { position: relative; overflow: hidden; display: inline-block; }
        .file-upload-wrapper input[type="file"] { font-size: 100px; position: absolute; left: 0; top: 0; opacity: 0; cursor: pointer; }
        .btn { background: transparent; color: var(--primary); border: 1px solid var(--primary); padding: 10px 20px; font-family: inherit; font-weight: bold; cursor: pointer; text-transform: uppercase; transition: all 0.3s; letter-spacing: 1px; display: inline-block; }
        .btn:hover { background: var(--primary); color: #000; }
        .file-name-display { color: #888; font-style: italic; font-size: 0.9em; }
        .terminal { background: var(--terminal-bg); border: 1px solid #333; border-top: 20px solid #222; padding: 15px; font-size: 0.9em; min-height: 200px; overflow-x: auto; position: relative; border-radius: 4px; white-space: pre-wrap; }
        .terminal::before { content: 'TERMINAL - CROSS-CHAIN OUTPUT'; position: absolute; top: -16px; left: 10px; color: #888; font-size: 10px; }
        .result-card { display: none; margin-top: 15px; padding: 20px; border-radius: 6px; text-align: center; }
        .result-card.valid { background: rgba(0,255,100,0.1); border: 1px solid #00cc66; }
        .result-card.revoked { background: rgba(255,150,0,0.1); border: 1px solid #ff9900; }
        .result-card.invalid { background: rgba(255,50,50,0.1); border: 1px solid #ff3333; }
        .result-icon { font-size: 48px; margin-bottom: 10px; }
        .result-title { font-size: 22px; font-weight: bold; margin-bottom: 8px; }
        .result-card.valid .result-title { color: #00ff66; }
        .result-card.revoked .result-title { color: #ff9900; }
        .result-card.invalid .result-title { color: #ff4444; }
        .result-desc { color: #888; font-size: 13px; }
        .result-cid { color: #555; font-size: 11px; margin-top: 8px; word-break: break-all; }
    </style>
</head>
<body>
    <div class="header"><h2>5G-A 异构跨域信任协同矩阵</h2></div>
    <div class="container">
        <div class="panel">
            <h3>阶段一：提取通感指纹并注入 L0 总线</h3>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="input-group">
                    <div class="file-upload-wrapper">
                        <button type="button" class="btn">选择物理载荷</button>
                        <input type="file" name="myFile" id="fileInput" onchange="updateFileName()">
                    </div>
                    <span id="fileName" class="file-name-display">等待接入数据...</span>
                </div>
                <button type="button" class="btn" style="width:100%;margin-top:10px;" onclick="uploadFile()">[ 执行跨域穿透 ]</button>
            </form>
        </div>
        <div class="panel">
            <h3>阶段二：查证底座状态机闭环回执</h3>
            <div class="input-group">
                <input type="text" id="queryId" placeholder="输入系统分配的 IPFS CID 进行溯源">
                <button class="btn" onclick="queryStatus()">[ 检索底层账本 ]</button>
            </div>
            <div class="result-card valid" id="cardValid">
                <div class="result-icon">✅</div>
                <div class="result-title">核验通过</div>
                <div class="result-desc">该证书已在 Hyperledger Fabric 权威账本中完成确权，内容完整性验证通过。</div>
                <div class="result-cid" id="cardValidCid"></div>
            </div>
            <div class="result-card revoked" id="cardRevoked">
                <div class="result-icon">⚠️</div>
                <div class="result-title">证书已吊销</div>
                <div class="result-desc">该证书已被颁发机构正式吊销，不再具有法律效力，请勿信任。</div>
                <div class="result-cid" id="cardRevokedCid"></div>
            </div>
            <div class="result-card invalid" id="cardInvalid">
                <div class="result-icon">❌</div>
                <div class="result-title">核验失败</div>
                <div class="result-desc">该证书未在权威账本中找到对应记录，可能为伪造或未经注册的凭证。</div>
                <div class="result-cid" id="cardInvalidCid"></div>
            </div>
        </div>
        <div class="panel">
            <h3>控制台核心输出</h3>
            <pre class="terminal" id="output">等待指令下发...</pre>
        </div>
    </div>
    <script>
        function updateFileName() {
            const input = document.getElementById('fileInput');
            const nameDisplay = document.getElementById('fileName');
            if (input.files.length > 0) {
                nameDisplay.innerText = "已锁定目标: " + input.files[0].name;
                nameDisplay.style.color = 'var(--primary)';
            } else {
                nameDisplay.innerText = '等待接入数据...';
                nameDisplay.style.color = '#888';
            }
        }
        function hideAllCards() {
            ['cardValid','cardRevoked','cardInvalid'].forEach(id => {
                document.getElementById(id).style.display = 'none';
            });
        }
        function showResultCard(status, cid) {
            hideAllCards();
            if (status === 'VALID') {
                document.getElementById('cardValid').style.display = 'block';
                document.getElementById('cardValidCid').textContent = 'CID: ' + cid;
            } else if (status === 'REVOKED') {
                document.getElementById('cardRevoked').style.display = 'block';
                document.getElementById('cardRevokedCid').textContent = 'CID: ' + cid;
            } else {
                document.getElementById('cardInvalid').style.display = 'block';
                document.getElementById('cardInvalidCid').textContent = 'CID: ' + cid;
            }
        }
        function uploadFile() {
            const fileInput = document.getElementById('fileInput');
            if (fileInput.files.length === 0) {
                document.getElementById('output').innerText = "[系统拦截] 未检测到物理载荷，请先选择文件。";
                return;
            }
            hideAllCards();
            let formData = new FormData(document.getElementById('uploadForm'));
            document.getElementById('output').innerText = "[系统] 正在计算 IPFS 载荷指纹...\n[系统] 正在向底层节点打包交易并抛出事件...\n\n";
            fetch('/upload', { method: 'POST', body: formData })
                .then(r => r.text())
                .then(t => {
                    document.getElementById('output').innerText += t;
                    let match = t.match(/ReqID: (Qm[a-zA-Z0-9]+)/);
                    if (match) document.getElementById('queryId').value = match[1];
                });
        }
        function queryStatus() {
            let id = document.getElementById('queryId').value.trim();
            if (!id) {
                document.getElementById('output').innerText = "[系统拦截] 缺少 CID，无法进行账本检索。";
                return;
            }
            hideAllCards();
            document.getElementById('output').innerText = "[系统] 正在穿透节点物理层，读取 L0 权威账本...\n\n";
            fetch('/fabric-query?id=' + id)
                .then(r => r.text())
                .then(t => {
                    document.getElementById('output').innerText += t + '\n\n';
                    if (t.startsWith('STATUS:VALID')) {
                        showResultCard('VALID', id);
                        document.getElementById('output').innerText += '--- L0 权威判决: 证书有效 ---';
                    } else if (t.startsWith('STATUS:REVOKED')) {
                        showResultCard('REVOKED', id);
                        document.getElementById('output').innerText += '--- L0 权威判决: 证书已吊销 ---';
                    } else if (t.startsWith('STATUS:NOTFOUND')) {
                        showResultCard('INVALID', id);
                        document.getElementById('output').innerText += '--- L0 权威判决: 证书不存在 ---';
                    } else {
                        showResultCard('INVALID', id);
                        document.getElementById('output').innerText += '--- 查询异常，请检查节点状态 ---';
                    }
                });
        }
    </script>
</body>
</html>`

func main() {
	fmt.Println("[Verifier UI] Dashboard started on port 8888")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/query", queryStatusHandler)
	http.HandleFunc("/fabric-query", queryFabricHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}
