# 🏛️ 异构双链跨链系统 - 核心架构与战略路线图

本仓库不仅是跨链预言机的代码库，更是我们整个《异构双链跨链存证系统》的战略大本营。本文档完整记录了系统的角色定义、5步流转全景，以及我们正在攻克的演进目标。

---

## 🎭 一、 角色定义（三位一体 + 物理层）

我们的架构由四个核心组件完美咬合：

1. **后端权威大本营 (Hyperledger Fabric)**
   - **职责**：运行 `pkicert.go` 和 `spbft.go`。只有信誉积分达标的节点，才有资格在这里颁发证书。
   - **特性**：数据绝对权威、隐私保护极高。
2. **前端高频业务链 (FISCO BCOS)**
   - **职责**：运行 `CertOracle.sol`。面对海量用户的查询请求。
   - **特性**：出块极快，适合作为业务流转的前端缓冲层。
3. **跨链枢纽中间件 (Go 预言机 / Chainlink)**
   - **职责**：7x24 小时运行，负责跨网络的数据搬运和密码学签名。
4. **物理存储层 (IPFS)**
   - **职责**：存储真实的证书文件（如 PDF、图片），仅将文件的特征值（哈希）上链，确保链上数据轻量化。

---

## 🔄 二、 全景数据流转 5 步曲
sequenceDiagram
    autonumber
    
    %% 定义参与角色
    participant User as 企业/高校 (前端)
    participant IPFS as IPFS 网络 (存储层)
    participant Fabric as Fabric (权威存证链)
    participant Oracle as 跨链预言机 (中继枢纽)
    participant FISCO as FISCO BCOS (前端业务链)

    %% 阶段一
    rect rgb(240, 248, 255)
    Note over User, Fabric: 阶段一：源头颁发与存证
    User->>IPFS: 上传毕业证书 PDF 等源文件
    IPFS-->>User: 返回文件唯一特征值 (ipfsHash)
    User->>Fabric: 调用 pkicert.go -> IssueCertificate(ipfsHash)
    Note right of Fabric: Fabric 确权，永久写入底层账本
    end

    %% 阶段二
    rect rgb(255, 245, 238)
    Note over User, FISCO: 阶段二：用户发起跨链查验
    User->>FISCO: 企业验证真伪，调用 CertOracle.sol 传入 ipfsHash
    FISCO->>FISCO: 检索本地账本 (查无此证)
    FISCO-->>Oracle: 智能合约抛出 Event: CertVerificationRequested
    end

    %% 阶段三
    rect rgb(240, 255, 240)
    Note over Oracle, FISCO: 阶段三：预言机捕获与路由
    Oracle->>Oracle: 7x24h 守护进程瞬间捕捉到 FISCO 事件
    Oracle->>Fabric: 解析 ipfsHash，跨网直连 Fabric gRPC 节点
    end

    %% 阶段四
    rect rgb(255, 250, 205)
    Note over Oracle, Fabric: 阶段四：权威核实（状态研判）
    Oracle->>Fabric: 调用 QueryCertificate(ipfsHash) 接口
    Fabric-->>Oracle: 返回证书状态 (IsValid) 与颁发者信誉积分
    Note right of Oracle: 逻辑研判：<br/>A: 存在 + 有效 + 信誉达标 -> 判定 True<br/>B: 不存在 或 已吊销 -> 判定 False
    end

    %% 阶段五
    rect rgb(230, 230, 250)
    Note over User, Oracle: 阶段五：结果回写与完美闭环
    Oracle->>FISCO: 预言机私钥签名，调用 verifyCert 强行回写 True/False
    Note left of FISCO: 状态跨链同步完成，数据落块
    FISCO-->>User: 监听到状态改变，前端展示“验证成功/失败”
    end
这是系统最核心的生命线，目前 MVP 版本已在物理层面上打通：

### 阶段一：源头颁发与存证（数据录入）
* **IPFS 文件上链**：某高校（Fabric 里的 Org1）生成毕业证书 PDF，上传至 IPFS，获得全球唯一指纹：`ipfsHash`（如 Qm...123）。
* **Fabric 权威铸造**：Org1 调用 Fabric 上的 `pkicert.go` 的 `IssueCertificate` 方法，将 `ipfsHash` 永久写入账本。Fabric 拥有唯一合法解释权。

### 阶段二：用户发起跨链查验（业务触发）
* **FISCO 侧查询**：企业拿到应聘者的证书哈希，通过网页调用 FISCO BCOS 的 `CertOracle.sol` 验证接口。
* **抛出跨链事件**：FISCO 账本查无此数据，智能合约抛出 Event（`CertVerificationRequested(ipfsHash)`），向外部预言机求助。

### 阶段三：预言机捕获与路由（核心中继）
* **精确监听**：挂在后台的预言机（Go 原生或 Chainlink 节点）瞬间捕捉到 FISCO 抛出的 Event。
* **跨网提取**：预言机解析出目标 `ipfsHash`，转身使用 Fabric SDK 连接服务链。

### 阶段四：权威核实（状态研判）
* **调用 Fabric**：预言机调用 Fabric 的 `QueryCertificate(ipfsHash)` 接口。
* **状态判定**：
  * **[情况 A - 有效]**：证书存在且 `IsValid == true`，且颁发者 S-PBFT 信誉正常 ➡️ 判定 `True`。
  * **[情况 B - 无效]**：查无此证，或已吊销 ➡️ 判定 `False`。

### 阶段五：结果回写与完美闭环（共识落定）
* **预言机签名回写**：预言机用私钥签名，调用 FISCO BCOS 的 `verifyCert` 方法，将 `True/False` 强行写回账本。
* **前端状态更新**：企业网页监听到 FISCO 状态变化，弹出真实性验证结果。

---

## 🗺️ 三、 终极演进路线图 (Checklist)

根据最初的项目规划，我们将逐步点亮以下科技树：

- [x] **1) Hyper Fabric 联盟链搭建**
- [x] **2) FISCO BCOS 联盟链搭建**
- [x] **3) 支持不同共识算法 (PBFT, DPoS, Raft/Kafka)**
- [x] **4) 智能合约应用 (已双端部署)**
- [ ] **5) 区块链浏览器可视化接入**
- [ ] **6) 预言机多形态演进：**
  - [x] 6.b) 链上数据获取 
  - [x] 6.c) 中心化预言机 (当前的 Go API/守护进程版)
  - [x] 6.d) 预言机 + IPFS 综合验证
  - [ ] **6.a) Chainlink 去中心化预言机实验 (下一步核心突破点)**
