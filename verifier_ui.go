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

// ── FISCO 控制台调用 ──────────────────────────────────────

func executeConsoleCmd(method string, args ...string) (string, error) {
	cmdArgs := []string{"console.sh", "call", Cfg.Fisco.ContractName, Cfg.Fisco.ContractAddr, method}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, fmt.Sprintf("\"%s\"", arg))
	}
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = Cfg.Fisco.ConsoleDir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ── Fabric 直查（L0 权威状态）────────────────────────────

type CertRecord struct {
	CertID    string `json:"certID"`
	Owner     string `json:"owner"`
	IPFSHash  string `json:"ipfsHash"`
	Status    string `json:"status"`
	IssuedAt  string `json:"issuedAt"`
	RevokedAt string `json:"revokedAt"`
}

func queryFabricDirect(certHash string) (string, *CertRecord) {
	cmd := exec.Command(Cfg.Fabric.PeerBin(), "chaincode", "query",
		"-C", Cfg.Fabric.Channel,
		"-n", Cfg.Fabric.Chaincode,
		"-c", fmt.Sprintf(`{"Args":["QueryCert", "%s"]}`, certHash))
	cmd.Env = append(os.Environ(), Cfg.Fabric.FabricEnv()...)

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if strings.Contains(outputStr, "NOT_FOUND") || strings.Contains(outputStr, "does not exist") {
			return "NOTFOUND", nil
		}
		return "ERROR", nil
	}

	var cert CertRecord
	if jsonErr := json.Unmarshal([]byte(outputStr), &cert); jsonErr != nil {
		return "ERROR", nil
	}

	switch cert.Status {
	case "VALID":
		return "VALID", &cert
	case "REVOKED":
		return "REVOKED", &cert
	default:
		return "NOTFOUND", nil
	}
}

// ── HTTP 处理器 ───────────────────────────────────────────

// 文件上传，计算 CID
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "文件读取失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "verify-*.tmp")
	if err != nil {
		http.Error(w, "临时文件创建失败", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	io.Copy(tmpFile, file)
	tmpFile.Close()

	cmd := exec.Command("ipfs", "add", "-Q", "-n", tmpFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "CID 计算失败: "+string(out), http.StatusInternalServerError)
		return
	}

	fingerprint := strings.TrimSpace(string(out))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"cid":"%s"}`, fingerprint)
}

// 触发跨链核验（写入 FISCO 合约，触发预言机）
func triggerVerifyHandler(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("id")
	if cid == "" {
		http.Error(w, "缺少 CID 参数", http.StatusBadRequest)
		return
	}

	sourceDID := "did:5g:base-station-001"
	targetDomain := "core-network-domain"
	outCmd, _ := executeConsoleCmd("requestDataAuth", cid, sourceDID, targetDomain, cid)

	if strings.Contains(outCmd, "Request ID already exists") {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"EXISTS","log":"该 CID 已有历史查验记录，直接读取账本状态"}`)
		return
	}

	if strings.Contains(outCmd, "transaction executed successfully") {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"OK","log":%s}`, jsonStr(outCmd))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ERROR","log":%s}`, jsonStr(outCmd))
}

// Fabric 直查（返回完整 JSON）
func fabricQueryHandler(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("id")
	if cid == "" {
		http.Error(w, "缺少 CID", http.StatusBadRequest)
		return
	}
	status, cert := queryFabricDirect(cid)
	w.Header().Set("Content-Type", "application/json")
	if cert != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"status": status,
			"cert":   cert,
		})
		w.Write(data)
	} else {
		fmt.Fprintf(w, `{"status":"%s","cert":null}`, status)
	}
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlPage)
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[VerifierUI] 配置加载失败: %v", err)
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/trigger", triggerVerifyHandler)
	http.HandleFunc("/api/fabric-query", fabricQueryHandler)
	fmt.Printf("[Verifier UI] 跨链核验系统已启动，端口 %s\n", Cfg.Ports.VerifierUI)
	log.Fatal(http.ListenAndServe(Cfg.Ports.VerifierUI, nil))
}

const htmlPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>CertChain Verifier</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'SF Pro Display','Segoe UI',Helvetica,sans-serif;background:#f5f5f7;min-height:100vh;color:#1d1d1f}
.nav{background:rgba(255,255,255,0.88);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);border-bottom:0.5px solid #d2d2d7;padding:0 32px;height:52px;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:100}
.nav-brand{display:flex;align-items:center;gap:10px}
.nav-icon{width:28px;height:28px;background:linear-gradient(135deg,#30d158,#00b4f0);border-radius:7px;display:flex;align-items:center;justify-content:center}
.nav-icon svg{width:15px;height:15px;fill:none;stroke:#fff;stroke-width:2.2;stroke-linecap:round;stroke-linejoin:round}
.nav-title{font-size:15px;font-weight:600;letter-spacing:-.3px}
.nav-sub{font-size:12px;color:#86868b;margin-left:6px}
.status-badge{font-size:11px;padding:4px 10px;border-radius:20px;background:#e8faf0;color:#1a7f3c;font-weight:500;display:flex;align-items:center;gap:5px}
.status-dot{width:6px;height:6px;border-radius:50%;background:#30d158}
.main{padding:28px 32px;max-width:860px;margin:0 auto}
/* 输入卡片 */
.verify-card{background:#fff;border-radius:16px;border:0.5px solid #e5e5ea;padding:28px 32px;margin-bottom:20px}
.verify-title{font-size:20px;font-weight:600;letter-spacing:-.3px;margin-bottom:5px}
.verify-sub{font-size:13px;color:#86868b;margin-bottom:22px;line-height:1.6}
.upload-row{display:flex;align-items:center;gap:10px;margin-bottom:14px}
.upload-btn{display:flex;align-items:center;gap:6px;padding:9px 16px;border-radius:8px;font-size:13px;font-weight:500;cursor:pointer;border:1px solid #d1d1d6;background:#fff;color:#1d1d1f;font-family:inherit;transition:all .15s}
.upload-btn:hover{background:#f5f5f7}
.upload-btn svg{width:14px;height:14px;stroke:#86868b;fill:none;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.upload-divider{font-size:12px;color:#c7c7cc}
.input-row{display:grid;grid-template-columns:1fr auto;gap:10px}
.cid-input{padding:12px 14px;border:1px solid #d1d1d6;border-radius:10px;font-size:13px;font-family:'SF Mono','Fira Code',monospace;color:#1d1d1f;background:#fff;outline:none;transition:border .15s;width:100%}
.cid-input:focus{border-color:#0071e3;box-shadow:0 0 0 3px rgba(0,113,227,.12)}
.btn-verify{padding:12px 28px;border-radius:10px;background:#0071e3;color:#fff;font-size:14px;font-weight:600;border:none;cursor:pointer;font-family:inherit;white-space:nowrap;transition:background .15s}
.btn-verify:hover{background:#0077ed}
.btn-verify:disabled{background:#c7c7cc;cursor:not-allowed}
/* 结果卡片 */
.result-wrap{margin-bottom:20px;display:none}
.result-card{border-radius:14px;padding:24px 28px;display:flex;align-items:center;gap:22px}
.result-card.valid{background:#e8faf0;border:1px solid #a3e6bc}
.result-card.revoked{background:#fff8ec;border:1px solid #ffd19a}
.result-card.invalid{background:#fff0f0;border:1px solid #ffb8b8}
.result-card.loading{background:#f5f5f7;border:1px solid #e5e5ea}
.result-icon{width:52px;height:52px;border-radius:13px;display:flex;align-items:center;justify-content:center;flex-shrink:0}
.result-card.valid .result-icon{background:#30d158}
.result-card.revoked .result-icon{background:#ff9f0a}
.result-card.invalid .result-icon{background:#ff3b30}
.result-card.loading .result-icon{background:#c7c7cc}
.result-icon svg{width:26px;height:26px;stroke:#fff;fill:none;stroke-width:2.5;stroke-linecap:round;stroke-linejoin:round}
.result-body{flex:1}
.result-title{font-size:18px;font-weight:600;letter-spacing:-.3px;margin-bottom:4px}
.result-card.valid .result-title{color:#1a7f3c}
.result-card.revoked .result-title{color:#b25c00}
.result-card.invalid .result-title{color:#c93030}
.result-card.loading .result-title{color:#86868b}
.result-desc{font-size:13px;color:#86868b;line-height:1.5;margin-bottom:10px}
.result-meta{display:flex;flex-wrap:wrap;gap:16px}
.meta-item{font-size:12px;color:#86868b}
.meta-item span{color:#1d1d1f;font-weight:500;margin-left:4px}
/* 详情卡片 */
.detail-card{background:#fff;border-radius:16px;border:0.5px solid #e5e5ea;overflow:hidden;margin-bottom:20px;display:none}
.detail-header{padding:15px 20px;border-bottom:0.5px solid #f2f2f7;display:flex;align-items:center;justify-content:space-between}
.detail-header-title{font-size:13px;font-weight:600;color:#1d1d1f}
.detail-grid{display:grid;grid-template-columns:1fr 1fr}
.detail-row{padding:14px 20px;border-bottom:0.5px solid #f5f5f7}
.detail-row:nth-last-child(-n+2){border-bottom:none}
.detail-label{font-size:11px;color:#86868b;font-weight:500;text-transform:uppercase;letter-spacing:.5px;margin-bottom:5px}
.detail-val{font-size:12.5px;color:#1d1d1f;font-family:'SF Mono','Fira Code',monospace;word-break:break-all;line-height:1.5}
.detail-val.green{color:#1a7f3c}
.detail-val.orange{color:#b25c00}
.detail-val.muted{color:#c7c7cc}
/* 终端 */
.terminal-card{background:#fff;border-radius:16px;border:0.5px solid #e5e5ea;overflow:hidden}
.terminal-header{padding:12px 20px;border-bottom:0.5px solid #f2f2f7;display:flex;align-items:center;gap:8px}
.tdots{display:flex;gap:6px}
.tdot{width:10px;height:10px;border-radius:50%}
.tdot.r{background:#ff5f57}.tdot.y{background:#ffbd2e}.tdot.g{background:#28c840}
.terminal-label{font-size:12px;color:#86868b;margin-left:4px}
.terminal-body{background:#f9f9fb;padding:16px 20px;font-family:'SF Mono','Fira Code',monospace;font-size:12px;min-height:90px;max-height:180px;overflow-y:auto}
.tline{display:flex;gap:8px;padding:2px 0;line-height:1.6}
.tprompt{color:#c7c7cc;flex-shrink:0}
.tok{color:#1a7f3c}.terr{color:#c93030}.twarn{color:#b25c00}.tinfo{color:#86868b}
input[type=file]{display:none}
</style>
</head>
<body>

<nav class="nav">
  <div class="nav-brand">
    <div class="nav-icon">
      <svg viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>
    </div>
    <span class="nav-title">CertChain Verifier</span>
    <span class="nav-sub">跨链证书核验系统</span>
  </div>
  <div class="status-badge">
    <span class="status-dot"></span>
    FISCO BCOS · Fabric 双链在线
  </div>
</nav>

<div class="main">

  <!-- 输入区 -->
  <div class="verify-card">
    <div class="verify-title">证书跨链核验</div>
    <div class="verify-sub">上传证书文件自动提取 CID，或直接粘贴 CID 进行查验。系统将通过 Chainlink 预言机跨链查询 Hyperledger Fabric 权威账本。</div>
    <div class="upload-row">
      <button class="upload-btn" onclick="document.getElementById('fileInput').click()">
        <svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
        上传文件自动提取 CID
      </button>
      <span class="upload-divider">或手动输入</span>
      <input type="file" id="fileInput" onchange="handleFileUpload(this.files[0])">
    </div>
    <div class="input-row">
      <input class="cid-input" id="cidInput" placeholder="输入 IPFS CID，例如 QmXf8enVq...">
      <button class="btn-verify" id="verifyBtn" onclick="doVerify()">执行核验</button>
    </div>
  </div>

  <!-- 结果卡片 -->
  <div class="result-wrap" id="resultWrap">
    <div class="result-card" id="resultCard">
      <div class="result-icon" id="resultIcon">
        <svg viewBox="0 0 24 24" id="resultIconSvg"></svg>
      </div>
      <div class="result-body">
        <div class="result-title" id="resultTitle"></div>
        <div class="result-desc" id="resultDesc"></div>
        <div class="result-meta" id="resultMeta"></div>
      </div>
    </div>
  </div>

  <!-- 链上详情 -->
  <div class="detail-card" id="detailCard">
    <div class="detail-header">
      <span class="detail-header-title">Fabric 账本记录</span>
    </div>
    <div class="detail-grid" id="detailGrid"></div>
  </div>

  <!-- 终端输出 -->
  <div class="terminal-card">
    <div class="terminal-header">
      <div class="tdots">
        <div class="tdot r"></div>
        <div class="tdot y"></div>
        <div class="tdot g"></div>
      </div>
      <span class="terminal-label">跨链预言机输出</span>
    </div>
    <div class="terminal-body" id="terminal">
      <div class="tline"><span class="tprompt">$</span><span class="tinfo">系统就绪，等待核验指令...</span></div>
    </div>
  </div>

</div>

<script>
function tlog(text, type) {
  const term = document.getElementById('terminal');
  const cls = {ok:'tok', err:'terr', warn:'twarn', info:'tinfo'}[type] || 'tinfo';
  const line = document.createElement('div');
  line.className = 'tline';
  line.innerHTML = '<span class="tprompt">$</span><span class="' + cls + '">' + text + '</span>';
  term.appendChild(line);
  term.scrollTop = term.scrollHeight;
}

function formatTime(iso) {
  if (!iso) return '—';
  return iso.replace('T',' ').replace('Z','').slice(0,16);
}

async function handleFileUpload(file) {
  if (!file) return;
  tlog('正在计算文件 CID: ' + file.name, 'info');
  const formData = new FormData();
  formData.append('myFile', file);
  try {
    const res = await fetch('/api/upload', { method:'POST', body:formData });
    const data = await res.json();
    document.getElementById('cidInput').value = data.cid;
    tlog('CID 提取成功: ' + data.cid, 'ok');
  } catch(e) {
    tlog('文件上传失败: ' + e.message, 'err');
  }
}

function showLoading() {
  const wrap = document.getElementById('resultWrap');
  const card = document.getElementById('resultCard');
  wrap.style.display = 'block';
  card.className = 'result-card loading';
  document.getElementById('resultTitle').textContent = '正在核验中...';
  document.getElementById('resultDesc').textContent = '正在通过预言机跨链查询 Fabric 权威账本';
  document.getElementById('resultMeta').innerHTML = '';
  document.getElementById('resultIconSvg').innerHTML = '<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>';
  document.getElementById('detailCard').style.display = 'none';
}

function showResult(status, cert, cid) {
  const card = document.getElementById('resultCard');
  const iconSvg = document.getElementById('resultIconSvg');
  const title = document.getElementById('resultTitle');
  const desc = document.getElementById('resultDesc');
  const meta = document.getElementById('resultMeta');

  if (status === 'VALID') {
    card.className = 'result-card valid';
    iconSvg.innerHTML = '<polyline points="20 6 9 17 4 12"/>';
    title.textContent = '核验通过';
    desc.textContent = '该证书已在 Hyperledger Fabric 权威账本中完成确权，内容完整性验证通过。';
    meta.innerHTML =
      '<div class="meta-item">颁发机构<span>' + (cert.owner||'—') + '</span></div>' +
      '<div class="meta-item">存证时间<span>' + formatTime(cert.issuedAt) + '</span></div>' +
      '<div class="meta-item">链上状态<span style="color:#1a7f3c">VALID</span></div>';
  } else if (status === 'REVOKED') {
    card.className = 'result-card revoked';
    iconSvg.innerHTML = '<path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>';
    title.textContent = '证书已吊销';
    desc.textContent = '该证书已被颁发机构正式吊销，不再具有法律效力，请勿信任该凭证。';
    meta.innerHTML =
      '<div class="meta-item">颁发机构<span>' + (cert.owner||'—') + '</span></div>' +
      '<div class="meta-item">吊销时间<span style="color:#b25c00">' + formatTime(cert.revokedAt) + '</span></div>' +
      '<div class="meta-item">链上状态<span style="color:#b25c00">REVOKED</span></div>';
  } else {
    card.className = 'result-card invalid';
    iconSvg.innerHTML = '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>';
    title.textContent = '核验失败';
    desc.textContent = '该证书未在 Fabric 权威账本中找到对应记录，可能为伪造凭证或未经注册的文件。';
    meta.innerHTML = '<div class="meta-item">CID<span style="font-family:monospace">' + cid.slice(0,20) + '...</span></div>';
  }

  // 详情面板
  if (cert) {
    const statusCls = cert.status === 'VALID' ? 'green' : 'orange';
    document.getElementById('detailGrid').innerHTML =
      '<div class="detail-row"><div class="detail-label">证书 CID</div><div class="detail-val">' + cert.certID + '</div></div>' +
      '<div class="detail-row"><div class="detail-label">颁发机构</div><div class="detail-val">' + cert.owner + '</div></div>' +
      '<div class="detail-row"><div class="detail-label">链上状态</div><div class="detail-val ' + statusCls + '">' + cert.status + '</div></div>' +
      '<div class="detail-row"><div class="detail-label">IPFS 哈希</div><div class="detail-val">' + cert.ipfsHash + '</div></div>' +
      '<div class="detail-row"><div class="detail-label">存证时间</div><div class="detail-val">' + (cert.issuedAt||'—') + '</div></div>' +
      '<div class="detail-row"><div class="detail-label">吊销时间</div><div class="detail-val ' + (cert.revokedAt ? 'orange' : 'muted') + '">' + (cert.revokedAt||'—') + '</div></div>';
    document.getElementById('detailCard').style.display = 'block';
  }
}

async function doVerify() {
  const cid = document.getElementById('cidInput').value.trim();
  if (!cid) { tlog('请输入 CID 后再执行核验', 'warn'); return; }

  document.getElementById('verifyBtn').disabled = true;
  showLoading();
  tlog('开始核验 CID: ' + cid.slice(0,24) + '...', 'info');

  try {
    // 第一步：直接查 Fabric 获取权威状态
    tlog('正在直查 Fabric L0 权威账本...', 'info');
    const fabRes = await fetch('/api/fabric-query?id=' + cid);
    const fabData = await fabRes.json();

    document.getElementById('resultWrap').style.display = 'block';

    if (fabData.status === 'VALID') {
      tlog('Fabric 账本查询完成 → STATUS: VALID', 'ok');
      showResult('VALID', fabData.cert, cid);
      // 同时触发跨链全流程（异步，不阻塞显示）
      tlog('正在触发预言机跨链全流程...', 'info');
      fetch('/api/trigger?id=' + cid).then(r => r.json()).then(d => {
        if (d.status === 'OK') tlog('预言机跨链闭环完成', 'ok');
        else if (d.status === 'EXISTS') tlog('已有历史记录，跨链流程跳过', 'info');
      });
    } else if (fabData.status === 'REVOKED') {
      tlog('Fabric 账本查询完成 → STATUS: REVOKED', 'warn');
      showResult('REVOKED', fabData.cert, cid);
    } else {
      tlog('Fabric 账本中未找到该 CID 记录', 'err');
      showResult('INVALID', null, cid);
    }
  } catch(e) {
    tlog('核验过程发生错误: ' + e.message, 'err');
    showResult('INVALID', null, cid);
  } finally {
    document.getElementById('verifyBtn').disabled = false;
  }
}

// 支持回车触发核验
document.getElementById('cidInput').addEventListener('keydown', e => {
  if (e.key === 'Enter') doVerify();
});
</script>
</body>
</html>`
