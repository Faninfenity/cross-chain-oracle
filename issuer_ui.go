package main

import (
	"bytes"
		"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ── Fabric 操作 ──────────────────────────────────────────

func issueToFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少指纹参数", http.StatusBadRequest)
		return
	}
	cmd := exec.Command("bash", "issue.sh", certHash)
	cmd.Dir = Cfg.Fabric.CliPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "ALREADY_EXISTS") || strings.Contains(string(out), "already exists") {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "ALREADY_EXISTS")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Fabric 写入失败: %v\n%s", err, string(out))
		}
		return
	}
	fmt.Fprintf(w, "OK")
}

func revokeFromFabricHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少指纹参数", http.StatusBadRequest)
		return
	}
	cmd := exec.Command("bash", "revoke.sh", certHash)
	cmd.Dir = Cfg.Fabric.CliPath
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		if strings.Contains(outStr, "证书不存在") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "NOT_FOUND")
		} else if strings.Contains(outStr, "已处于吊销状态") {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "ALREADY_REVOKED")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "ERROR: %s", outStr)
		}
		return
	}
	fmt.Fprintf(w, "OK")
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

	req, _ := http.NewRequest("POST", Cfg.IPFS.AddURL(), body)
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

func checkStatusHandler(w http.ResponseWriter, r *http.Request) {
	certHash := r.URL.Query().Get("id")
	if certHash == "" {
		http.Error(w, "缺少参数", http.StatusBadRequest)
		return
	}
	cmd := exec.Command(Cfg.Fabric.PeerBin(), "chaincode", "query",
		"-C", Cfg.Fabric.Channel,
		"-n", Cfg.Fabric.Chaincode,
		"-c", fmt.Sprintf(`{"Args":["QueryCert","%s"]}`, certHash))
	cmd.Env = append(os.Environ(), Cfg.Fabric.FabricEnv()...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "NOTFOUND")
		return
	}
	fmt.Fprintf(w, "%s", strings.TrimSpace(string(out)))
}

type CertRecord struct {
	CertID    string `json:"certID"`
	Owner     string `json:"owner"`
	IPFSHash  string `json:"ipfsHash"`
	Status    string `json:"status"`
	IssuedAt  string `json:"issuedAt"`
	RevokedAt string `json:"revokedAt"`
}

func getAllCertsHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command(Cfg.Fabric.PeerBin(), "chaincode", "query",
		"-C", Cfg.Fabric.Channel,
		"-n", Cfg.Fabric.Chaincode,
		"-c", `{"Args":["GetAllCerts"]}`)
	cmd.Env = append(os.Environ(), Cfg.Fabric.FabricEnv()...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "[]")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", strings.TrimSpace(string(out)))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlPage)
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("[IssuerUI] 配置加载失败: %v", err)
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/upload", uploadToIPFSHandler)
	http.HandleFunc("/api/issue", issueToFabricHandler)
	http.HandleFunc("/api/revoke", revokeFromFabricHandler)
	http.HandleFunc("/api/check", checkStatusHandler)
	http.HandleFunc("/api/certs", getAllCertsHandler)
	fmt.Printf("源头存证系统已启动，端口 %s\n", Cfg.Ports.IssuerUI)
	log.Fatal(http.ListenAndServe(Cfg.Ports.IssuerUI, nil))
}

// 当前时间（供模板使用）
var _ = time.Now

const htmlPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>CertChain Console</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'SF Pro Display','Segoe UI',Helvetica,sans-serif;background:#f5f5f7;min-height:100vh;color:#1d1d1f}
/* 导航栏 */
.nav{background:rgba(255,255,255,0.88);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);border-bottom:0.5px solid #d2d2d7;padding:0 32px;height:52px;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:100}
.nav-brand{display:flex;align-items:center;gap:10px}
.nav-icon{width:28px;height:28px;background:linear-gradient(135deg,#0071e3,#34aadc);border-radius:7px;display:flex;align-items:center;justify-content:center}
.nav-icon svg{width:15px;height:15px;fill:none;stroke:#fff;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.nav-title{font-size:15px;font-weight:600;letter-spacing:-.3px}
.nav-right{display:flex;align-items:center;gap:10px}
.status-badge{font-size:11px;padding:4px 10px;border-radius:20px;background:#e8f4fd;color:#0071e3;font-weight:500;display:flex;align-items:center;gap:5px}
.status-dot{width:6px;height:6px;border-radius:50%;background:#30d158}
.btn-refresh{padding:6px 14px;border-radius:20px;font-size:13px;font-weight:500;cursor:pointer;border:1px solid #c7c7cc;background:transparent;color:#1d1d1f;font-family:inherit;transition:background .15s}
.btn-refresh:hover{background:#f5f5f7}
/* 主体 */
.main{padding:28px 32px;max-width:1280px;margin:0 auto}
/* 统计卡片 */
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:24px}
.stat-card{background:#fff;border-radius:12px;padding:18px 20px;border:0.5px solid #e5e5ea}
.stat-label{font-size:12px;color:#86868b;font-weight:500;margin-bottom:6px}
.stat-num{font-size:28px;font-weight:600;letter-spacing:-.5px}
.stat-num.blue{color:#0071e3}
.stat-num.green{color:#30d158}
.stat-num.red{color:#ff3b30}
.stat-num.gray{color:#1d1d1f}
/* 内容区 */
.content{display:grid;grid-template-columns:360px 1fr;gap:16px}
/* 左侧面板 */
.side-panel{background:#fff;border-radius:16px;border:0.5px solid #e5e5ea;overflow:hidden}
.side-section{padding:20px}
.side-section+.side-section{border-top:0.5px solid #f2f2f7}
.section-title{font-size:13px;font-weight:600;color:#1d1d1f;margin-bottom:14px}
/* 上传区 */
.drop-zone{border:1.5px dashed #d1d1d6;border-radius:10px;padding:24px 16px;text-align:center;cursor:pointer;transition:all .2s;background:#fafafa}
.drop-zone:hover,.drop-zone.dragover{border-color:#0071e3;background:#f0f8ff}
.drop-icon{width:38px;height:38px;background:#f2f2f7;border-radius:9px;margin:0 auto 10px;display:flex;align-items:center;justify-content:center}
.drop-icon svg{width:18px;height:18px;fill:none;stroke:#86868b;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}
.drop-primary{font-size:14px;font-weight:500;color:#1d1d1f;margin-bottom:3px}
.drop-secondary{font-size:12px;color:#86868b}
.drop-secondary span{color:#0071e3;cursor:pointer}
/* CID 展示 */
.cid-box{background:#f5f5f7;border-radius:9px;padding:11px 13px;margin-top:12px;display:none}
.cid-box-label{font-size:10px;color:#86868b;font-weight:600;text-transform:uppercase;letter-spacing:.6px;margin-bottom:4px}
.cid-box-val{font-family:'SF Mono','Fira Code','Courier New',monospace;font-size:11.5px;color:#0071e3;word-break:break-all;line-height:1.6}
/* 按钮 */
.btn-issue{width:100%;padding:12px;border-radius:10px;font-size:14px;font-weight:600;border:none;cursor:pointer;font-family:inherit;transition:all .2s;margin-top:12px;letter-spacing:-.2px;background:#c7c7cc;color:#fff}
.btn-issue.ready{background:#0071e3;color:#fff}
.btn-issue.ready:hover{background:#0077ed}
/* 吊销区 */
.revoke-row{display:flex;gap:8px}
.revoke-input{flex:1;padding:10px 12px;border:1px solid #d1d1d6;border-radius:8px;font-size:12px;font-family:'SF Mono','Fira Code',monospace;color:#1d1d1f;background:#fff;outline:none;transition:border .15s}
.revoke-input:focus{border-color:#0071e3;box-shadow:0 0 0 3px rgba(0,113,227,.12)}
.btn-revoke{padding:10px 14px;border-radius:8px;background:#fff0f0;border:1px solid #ffd0d0;color:#ff3b30;font-size:12px;font-weight:600;cursor:pointer;font-family:inherit;white-space:nowrap;transition:background .15s}
.btn-revoke:hover{background:#ffe5e5}
.revoke-warn{font-size:11px;color:#86868b;margin-top:7px;line-height:1.5}
/* 日志框 */
.log-box{background:#f9f9fb;border:0.5px solid #e5e5ea;border-radius:9px;padding:10px 12px;height:100px;overflow-y:auto;margin-top:12px}
.log-row{display:flex;align-items:flex-start;gap:8px;padding:2.5px 0;font-size:12px}
.log-dot{width:6px;height:6px;border-radius:50%;margin-top:4px;flex-shrink:0}
.log-dot.ok{background:#30d158}
.log-dot.err{background:#ff3b30}
.log-dot.warn{background:#ff9f0a}
.log-dot.info{background:#c7c7cc}
.log-text{color:#1d1d1f;font-family:'SF Mono','Fira Code',monospace;flex:1;line-height:1.5}
.log-time{color:#c7c7cc;font-size:11px;font-family:'SF Mono',monospace;white-space:nowrap;padding-top:1px}
/* 右侧台账 */
.table-panel{background:#fff;border-radius:16px;border:0.5px solid #e5e5ea;overflow:hidden;display:flex;flex-direction:column}
.table-top{display:flex;align-items:center;justify-content:space-between;padding:18px 20px;border-bottom:0.5px solid #f2f2f7;flex-shrink:0}
.table-top-left{display:flex;align-items:center;gap:12px}
.table-top-title{font-size:15px;font-weight:600}
.table-count{font-size:12px;color:#86868b;background:#f2f2f7;padding:3px 9px;border-radius:20px}
.filter-tabs{display:flex;gap:2px;background:#f2f2f7;border-radius:8px;padding:2px}
.tab{padding:5px 13px;border-radius:6px;font-size:12px;font-weight:500;color:#86868b;cursor:pointer;border:none;background:transparent;font-family:inherit;transition:all .15s}
.tab.active{background:#fff;color:#1d1d1f;box-shadow:0 1px 2px rgba(0,0,0,.08)}
.table-wrap{overflow-y:auto;flex:1}
table{width:100%;border-collapse:collapse}
th{padding:10px 20px;text-align:left;font-size:11px;font-weight:600;color:#86868b;background:#fafafa;letter-spacing:.4px;position:sticky;top:0;border-bottom:0.5px solid #f2f2f7}
td{padding:13px 20px;font-size:13px;border-bottom:0.5px solid #f5f5f7;vertical-align:middle}
tr:last-child td{border-bottom:none}
tr:hover td{background:#fafafa}
.td-cid{font-family:'SF Mono','Fira Code',monospace;font-size:11.5px;color:#636366;max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;cursor:default}
.badge{display:inline-flex;align-items:center;gap:5px;padding:4px 10px;border-radius:20px;font-size:11.5px;font-weight:500}
.badge.valid{background:#e8faf0;color:#1a7f3c}
.badge.revoked{background:#fff0f0;color:#c93030}
.bdot{width:5px;height:5px;border-radius:50%}
.badge.valid .bdot{background:#30d158}
.badge.revoked .bdot{background:#ff3b30}
.td-time{font-size:12px;color:#86868b;font-family:'SF Mono',monospace}
.td-actions{display:flex;gap:6px}
.act-btn{padding:5px 11px;border-radius:6px;font-size:12px;font-weight:500;cursor:pointer;border:none;font-family:inherit;transition:background .15s}
.act-view{background:#f0f8ff;color:#0071e3}
.act-view:hover{background:#dff0ff}
.act-revoke{background:#fff0f0;color:#ff3b30}
.act-revoke:hover{background:#ffe5e5}
.empty-state{text-align:center;padding:60px 20px;color:#86868b;font-size:14px}
input[type=file]{display:none}
</style>
</head>
<body>

<nav class="nav">
  <div class="nav-brand">
    <div class="nav-icon">
      <svg viewBox="0 0 24 24"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
    </div>
    <span class="nav-title">CertChain Console</span>
  </div>
  <div class="nav-right">
    <div class="status-badge">
      <span class="status-dot"></span>
      Fabric 已连接
    </div>
    <button class="btn-refresh" onclick="loadCerts()">刷新台账</button>
  </div>
</nav>

<div class="main">
  <div class="stats">
    <div class="stat-card">
      <div class="stat-label">链上证书总数</div>
      <div class="stat-num blue" id="statTotal">—</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">有效证书</div>
      <div class="stat-num green" id="statValid">—</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">已吊销</div>
      <div class="stat-num red" id="statRevoked">—</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">IPFS 节点</div>
      <div class="stat-num gray" style="font-size:16px;padding-top:6px">port 5001</div>
    </div>
  </div>

  <div class="content">
    <!-- 左侧操作面板 -->
    <div class="side-panel">
      <!-- 存证 -->
      <div class="side-section">
        <div class="section-title">新建存证</div>
        <div class="drop-zone" id="dropZone">
          <div class="drop-icon">
            <svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </div>
          <div class="drop-primary">拖拽文件至此上传</div>
          <div class="drop-secondary">或 <span onclick="document.getElementById('fileInput').click()">点击选择文件</span> · 支持任意格式</div>
          <input type="file" id="fileInput">
        </div>
        <div class="cid-box" id="cidBox">
          <div class="cid-box-label">IPFS Content Identifier</div>
          <div class="cid-box-val" id="cidVal"></div>
        </div>
        <button class="btn-issue" id="issueBtn" onclick="issueToFabric()">将指纹写入 Fabric 账本</button>
      </div>

      <!-- 吊销 -->
      <div class="side-section">
        <div class="section-title">证书吊销</div>
        <div class="revoke-row">
          <input class="revoke-input" id="revokeInput" placeholder="输入证书 CID...">
          <button class="btn-revoke" onclick="revokeFromFabric()">吊销</button>
        </div>
        <div class="revoke-warn">⚠ 吊销操作上链后不可逆，请确认 CID 后执行</div>
      </div>

      <!-- 日志 -->
      <div class="side-section">
        <div class="section-title">操作日志</div>
        <div class="log-box" id="logBox">
          <div class="log-row">
            <span class="log-dot info"></span>
            <span class="log-text">系统就绪，等待操作</span>
            <span class="log-time" id="initTime"></span>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧台账 -->
    <div class="table-panel">
      <div class="table-top">
        <div class="table-top-left">
          <span class="table-top-title">证书台账</span>
          <span class="table-count" id="tableCount">加载中...</span>
        </div>
        <div class="filter-tabs">
          <button class="tab active" onclick="filterCerts('all',this)">全部</button>
          <button class="tab" onclick="filterCerts('VALID',this)">有效</button>
          <button class="tab" onclick="filterCerts('REVOKED',this)">已吊销</button>
        </div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>证书 CID</th>
              <th>颁发机构</th>
              <th>状态</th>
              <th>存证时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody id="certTable">
            <tr><td colspan="5" class="empty-state">正在从 Fabric 账本加载数据...</td></tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>

<script>
let allCerts = [];
let currentCID = null;

function nowTime() {
  return new Date().toTimeString().slice(0,8);
}

document.getElementById('initTime').textContent = nowTime();

function appendLog(text, type) {
  const box = document.getElementById('logBox');
  const row = document.createElement('div');
  row.className = 'log-row';
  const dotClass = {ok:'ok', err:'err', warn:'warn', info:'info'}[type] || 'info';
  row.innerHTML = '<span class="log-dot ' + dotClass + '"></span>' +
    '<span class="log-text">' + text + '</span>' +
    '<span class="log-time">' + nowTime() + '</span>';
  box.appendChild(row);
  box.scrollTop = box.scrollHeight;
}

// 文件上传
const dropZone = document.getElementById('dropZone');
const fileInput = document.getElementById('fileInput');

dropZone.addEventListener('dragover', e => { e.preventDefault(); dropZone.classList.add('dragover'); });
dropZone.addEventListener('dragleave', () => dropZone.classList.remove('dragover'));
dropZone.addEventListener('drop', e => { e.preventDefault(); dropZone.classList.remove('dragover'); handleFile(e.dataTransfer.files[0]); });
dropZone.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', () => handleFile(fileInput.files[0]));

async function handleFile(file) {
  if (!file) return;
  appendLog('正在上传至 IPFS: ' + file.name, 'info');
  const formData = new FormData();
  formData.append('file', file);
  try {
    const res = await fetch('/api/upload', { method: 'POST', body: formData });
    const data = await res.json();
    currentCID = data.Hash;
    document.getElementById('cidVal').textContent = currentCID;
    document.getElementById('cidBox').style.display = 'block';
    document.getElementById('revokeInput').value = currentCID;
    const btn = document.getElementById('issueBtn');
    btn.classList.add('ready');
    appendLog('IPFS 落块成功: ' + currentCID.slice(0,20) + '...', 'ok');
  } catch(e) {
    appendLog('IPFS 上传失败: ' + e.message, 'err');
  }
}

async function issueToFabric() {
  if (!currentCID) return;
  appendLog('正在写入 Fabric 账本...', 'info');
  try {
    const res = await fetch('/api/issue?id=' + currentCID);
    if (res.status === 200) {
      appendLog('存证成功: ' + currentCID.slice(0,20) + '...', 'ok');
      setTimeout(loadCerts, 1500);
    } else if (res.status === 409) {
      const checkRes = await fetch('/api/check?id=' + currentCID);
      const checkText = await checkRes.text();
      if (checkText.includes('"status":"REVOKED"')) {
        appendLog('该证书已被吊销，不可重新颁发', 'err');
      } else {
        appendLog('该证书已存在于账本，无需重复颁发', 'warn');
      }
    } else {
      const t = await res.text();
      appendLog('写入失败: ' + t.slice(0,50), 'err');
    }
  } catch(e) {
    appendLog('网络错误: ' + e.message, 'err');
  }
}

async function revokeFromFabric() {
  const cid = document.getElementById('revokeInput').value.trim();
  if (!cid) { appendLog('请输入待吊销证书的 CID', 'warn'); return; }
  if (!confirm('确认吊销证书？\n\nCID: ' + cid + '\n\n此操作不可逆。')) return;
  appendLog('正在执行吊销: ' + cid.slice(0,20) + '...', 'warn');
  try {
    const res = await fetch('/api/revoke?id=' + cid);
    if (res.status === 200) {
      appendLog('吊销成功: ' + cid.slice(0,20) + '...', 'ok');
      setTimeout(loadCerts, 1500);
    } else if (res.status === 404) {
      appendLog('证书不存在，请确认 CID', 'err');
    } else if (res.status === 409) {
      appendLog('该证书已处于吊销状态', 'warn');
    } else {
      appendLog('吊销失败，请检查链码状态', 'err');
    }
  } catch(e) {
    appendLog('网络错误: ' + e.message, 'err');
  }
}

// 吊销快捷操作（从台账行触发）
function quickRevoke(cid) {
  document.getElementById('revokeInput').value = cid;
  revokeFromFabric();
}

function formatTime(iso) {
  if (!iso) return '—';
  return iso.replace('T', ' ').replace('Z', '').slice(0,16);
}

let currentFilter = 'all';

function filterCerts(filter, el) {
  currentFilter = filter;
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  el.classList.add('active');
  renderTable(allCerts);
}

function renderTable(certs) {
  const filtered = currentFilter === 'all' ? certs : certs.filter(c => c.status === currentFilter);
  const tbody = document.getElementById('certTable');
  document.getElementById('tableCount').textContent = filtered.length + ' 条记录';

  if (filtered.length === 0) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty-state">暂无记录</td></tr>';
    return;
  }

  tbody.innerHTML = filtered.map(c => {
    const isValid = c.status === 'VALID';
    const badge = isValid
      ? '<span class="badge valid"><span class="bdot"></span>有效</span>'
      : '<span class="badge revoked"><span class="bdot"></span>已吊销</span>';
    const shortCID = c.certID.length > 24 ? c.certID.slice(0,12) + '...' + c.certID.slice(-8) : c.certID;
    const actions = isValid
      ? '<button class="act-btn act-revoke" onclick="quickRevoke(\'' + c.certID + '\')">吊销</button>'
      : '';
    return '<tr>' +
      '<td class="td-cid" title="' + c.certID + '">' + shortCID + '</td>' +
      '<td style="font-size:13px">' + c.owner + '</td>' +
      '<td>' + badge + '</td>' +
      '<td class="td-time">' + formatTime(c.issuedAt) + '</td>' +
      '<td><div class="td-actions">' + actions + '</div></td>' +
      '</tr>';
  }).join('');
}

function updateStats(certs) {
  const total = certs.length;
  const valid = certs.filter(c => c.status === 'VALID').length;
  const revoked = certs.filter(c => c.status === 'REVOKED').length;
  document.getElementById('statTotal').textContent = total;
  document.getElementById('statValid').textContent = valid;
  document.getElementById('statRevoked').textContent = revoked;
}

async function loadCerts() {
  try {
    const res = await fetch('/api/certs');
    const data = await res.json();
    allCerts = Array.isArray(data) ? data : [];
    // 按时间倒序排列
    allCerts.sort((a,b) => b.issuedAt.localeCompare(a.issuedAt));
    updateStats(allCerts);
    renderTable(allCerts);
  } catch(e) {
    document.getElementById('certTable').innerHTML =
      '<tr><td colspan="5" class="empty-state">加载失败，请检查 Fabric 节点状态</td></tr>';
  }
}

// 页面加载时自动拉取数据
loadCerts();
</script>
</body>
</html>`
