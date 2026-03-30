package main

import (
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
	
	// 向底层发起写入请求
	outCmd, _ := executeConsoleCmd("requestDataAuth", reqId, sourceDID, targetDomain, fingerprint)
	
	// 🎯 核心演示优化：智能拦截“防重放报错”
	if strings.Contains(outCmd, "Request ID already exists") {
		fmt.Fprintf(w, "[智能路由] 拦截到重复请求！\n\n系统检测到该物理载荷 (CID) 已在 L0 账本中存在历史共识记录。\nReqID: %s\n\n已为您省去重复的跨域预言机 Gas 消耗！\n请直接在下方点击【检索底层账本】查看历史权威判决。", reqId)
		return
	}

	// 如果是全新文件，则正常打印回执并触发事件总线
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>5G-A 异构跨域信任协同矩阵</title>
    <style>
        :root {
            --primary: #00ffcc;
            --bg-dark: #0a0a0f;
            --panel-bg: rgba(16, 20, 30, 0.85);
            --border-color: rgba(0, 255, 204, 0.3);
            --text-main: #e0e0e0;
            --terminal-bg: #050505;
        }
        body {
            background-color: var(--bg-dark);
            background-image:
                linear-gradient(rgba(0, 255, 204, 0.03) 1px, transparent 1px),
                linear-gradient(90deg, rgba(0, 255, 204, 0.03) 1px, transparent 1px);
            background-size: 30px 30px;
            color: var(--text-main);
            font-family: 'Consolas', 'Courier New', monospace;
            margin: 0;
            padding: 40px;
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            text-transform: uppercase;
            letter-spacing: 2px;
            text-shadow: 0 0 10px rgba(0,255,204,0.5);
            color: var(--primary);
            border-bottom: 1px solid var(--primary);
            padding-bottom: 15px;
            width: 80%;
            max-width: 900px;
        }
        .container {
            width: 80%;
            max-width: 900px;
            display: grid;
            gap: 25px;
        }
        .panel {
            background: var(--panel-bg);
            border: 1px solid var(--border-color);
            box-shadow: 0 0 15px rgba(0,255,204,0.05);
            border-radius: 4px;
            padding: 25px;
            position: relative;
            overflow: hidden;
        }
        .panel::before {
            content: '';
            position: absolute;
            top: 0; left: 0;
            width: 4px; height: 100%;
            background: var(--primary);
        }
        h3 {
            margin-top: 0;
            color: #fff;
            font-size: 1.2em;
            letter-spacing: 1px;
            border-bottom: 1px dashed #333;
            padding-bottom: 10px;
        }
        .input-group {
            margin: 20px 0;
            display: flex;
            gap: 15px;
            align-items: center;
        }
        input[type="text"] {
            flex: 1;
            background: rgba(0,0,0,0.5);
            border: 1px solid var(--border-color);
            color: var(--primary);
            padding: 12px;
            font-family: inherit;
            outline: none;
            transition: border 0.3s;
        }
        input[type="text"]:focus {
            border-color: var(--primary);
            box-shadow: 0 0 8px rgba(0,255,204,0.3);
        }
        .file-upload-wrapper {
            position: relative;
            overflow: hidden;
            display: inline-block;
        }
        .file-upload-wrapper input[type="file"] {
            font-size: 100px;
            position: absolute;
            left: 0;
            top: 0;
            opacity: 0;
            cursor: pointer;
        }
        .btn {
            background: transparent;
            color: var(--primary);
            border: 1px solid var(--primary);
            padding: 10px 20px;
            font-family: inherit;
            font-weight: bold;
            cursor: pointer;
            text-transform: uppercase;
            transition: all 0.3s ease;
            letter-spacing: 1px;
            display: inline-block;
        }
        .btn:hover {
            background: var(--primary);
            color: #000;
            box-shadow: 0 0 15px rgba(0,255,204,0.4);
        }
        .terminal {
            background: var(--terminal-bg);
            border: 1px solid #333;
            border-top: 20px solid #222;
            padding: 15px;
            color: #0f0;
            font-size: 0.9em;
            min-height: 200px;
            overflow-x: auto;
            position: relative;
            border-radius: 4px;
            white-space: pre-wrap;
        }
        .terminal::before {
            content: 'TERMINAL - L0 NODE RESPONSE';
            position: absolute;
            top: -16px; left: 10px;
            color: #888;
            font-size: 10px;
        }
        .file-name-display {
            color: #888;
            font-style: italic;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>5G-A 异构跨域信任协同矩阵</h2>
    </div>
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
                <button type="button" class="btn" style="width: 100%; margin-top: 10px;" onclick="uploadFile()">[ 执行跨域穿透 ]</button>
            </form>
        </div>

        <div class="panel">
            <h3>阶段二：查证底座状态机闭环回执</h3>
            <div class="input-group">
                <input type="text" id="queryId" placeholder="输入系统分配的 IPFS CID 进行溯源">
                <button class="btn" onclick="queryStatus()">[ 检索底层账本 ]</button>
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

        function uploadFile() {
            const fileInput = document.getElementById('fileInput');
            if (fileInput.files.length === 0) {
                document.getElementById('output').innerText = "[系统拦截] 未检测到物理载荷，请先选择文件。";
                return;
            }
            let formData = new FormData(document.getElementById('uploadForm'));
            document.getElementById('output').innerText = "[系统] 正在计算 IPFS 载荷指纹...\n[系统] 正在向底层节点打包交易并抛出事件...\n\n";
            fetch('/upload', { method: 'POST', body: formData })
                .then(r => r.text())
                .then(t => {
                    document.getElementById('output').innerText += t;
                    let match = t.match(/ReqID: (Qm[a-zA-Z0-9]+)/);
                    if (match) {
                        document.getElementById('queryId').value = match[1];
                    }
                });
        }

        function queryStatus() {
            let id = document.getElementById('queryId').value;
            if (!id) {
                document.getElementById('output').innerText = "[系统拦截] 缺少 ReqID，无法进行账本检索。";
                return;
            }
            document.getElementById('output').innerText = "[系统] 正在穿透节点物理层，读取账本映射...\n\n";
            fetch('/query?id=' + id)
                .then(r => r.text())
                .then(t => document.getElementById('output').innerText += t);
        }
    </script>
</body>
</html>`
	fmt.Fprint(w, html)
}

func main() {
	fmt.Println("[Verifier UI] Dashboard started on port 8888")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/query", queryStatusHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

