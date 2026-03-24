# Cross-Chain Oracle: 异构区块链跨链确权预言机系统

**项目名称**: cross-chain-oracle
**开发者**: Faninfenity
**核心技术栈**: Go, C++, Solidity, Docker, Hyperledger Fabric, FISCO BCOS, Chainlink, IPFS

---

## 1. 项目背景与设计理念 (Project Background)

在多链并存的联盟链生态中，数据孤岛问题日益凸显。传统的单链存证系统无法满足跨域信任传递的需求。本项目提出并实现了一种跨异构区块链（Hyperledger Fabric 与 FISCO BCOS）的分布式确权与核验架构。

系统以 IPFS（星际文件系统）作为去中心化底层存储，以 Hyperledger Fabric 作为高隐私、强管控的源头确权账本，以 FISCO BCOS 作为面向公众的快速核验侧链。通过引入 Chainlink 去中心化预言机集群，实现了异构账本间的状态安全监听、数据物理穿透与异步确权回写，彻底打通了跨链信任闭环。

## 2. 系统核心架构与微服务矩阵 (Core Architecture)

本系统的跨链交互并非简单的接口调用，而是基于区块链状态机的物理级握手，由以下核心组件构成：

| 微服务组件 | 进程 / 端口 | 核心职责与工程描述 |
| :--- | :--- | :--- |
| **源头存证大屏** | `issuer_ui` (:8889) | 负责文件哈希提取，并将数字资产指纹双写至 IPFS 与 Fabric 源头账本。 |
| **跨链查证大屏** | `verifier_ui` (:8888) | 提供核验入口，向 FISCO BCOS 抛出核验事件并轮询最终回写状态。 |
| **底层嗅探中枢** | `auto_trigger` (后台) | 基于 Go SDK 深度定制。直接嗅探区块事件，防死循环拦截，具备 ABI 动态解码能力。 |
| **预言机调度网关** | `Chainlink Node` (:6688) | 接收 Webhook 触发，通过有向无环图 (DAG) 工作流调度外部适配器集群。 |
| **Fabric 适配器** | `fabric_adapter` (:8081) | 作为 External Adapter，调用 Go Gateway 穿透 Fabric 网关提取真实状态。 |
| **FISCO 回写中枢** | `fisco_writer` (:8082) | 将穿透提取的数据重新 ABI 编码，签名发送至 FISCO BCOS，完成状态机闭环。 |

## 3. 跨链生命周期数据流转 (Data Flow)

1. [用户] 在 Verifier UI 上传需核验的文件。
2. [Verifier] 计算文件哈希，调用 FISCO BCOS 智能合约抛出 `CertVerificationRequested` 事件。
3. [Listener] 嗅探中枢拦截交易，提取 Hex Input 数据，动态反编译出真实的 IPFS CID。
4. [Listener] 通过认证 Webhook 将 CID 注入 Chainlink 节点的工作流 (Job Run)。
5. [Chainlink] 调度 Fabric Adapter，将 CID 发送至 8081 端口。
6. [Adapter] 载入证书连接 Peer 节点，执行链码查询，返回布尔值确权结果。
7. [Chainlink] 接收真实结果，触发 FISCO Writer 任务 (8082 端口)。
8. [Writer] 构建回写交易，调用智能合约 `fulfillVerification` 物理落块。
9. [Verifier] 大屏轮询检测到账本状态变更，向前端反馈最终核验结果。

## 4. 核心技术攻坚与工程亮点 (Technical Highlights)

* **EVM ABI 跨 Word 边界的动态自适应解析**: 在以太坊虚拟机底层，数据严格按 32 字节对齐。针对 46 字节长度的 IPFS CID，本系统重构了底层 Hex 解析引擎，通过提取偏移量与长度位，实现了对跨越 Word 边界的超长字符串精准剥离，确保跨链通信数据绝对保真。
* **异构网络协议栈平滑桥接**: 创新性引入双向外置适配器架构，将 Fabric 的 gRPC/Protobuf 与 FISCO BCOS 的 JSON-RPC 复杂交互下沉至独立的 Go 微服务，确保 Chainlink 核心节点的高可用性。
* **工业级进程生命周期管理**: 摈弃脆弱的开发级拉起，设计了基于 Bash 的全自动编译与 PID 追踪启停脚本，实现了跨链微服务集群的精准守护与平滑退出 (Graceful Shutdown)。

## 5. 极客部署指南 (Deployment)

本项目自带工程级防僵尸启停序列，确保多节点环境下的自动化流转。

**一键点火序列 (全量预编译与后台拉起):**
```bash
chmod +x start_all.sh
./start_all.sh

##6. 演进路线图 (Roadmap)
[x] 高优先级: 解决 Chainlink Webhook 调用的 401 Basic Auth 认证拦截问题。

[x] 高优先级: 在 adapter.go 中剥离 Mock 逻辑，接入真实的 Fabric Go SDK 账本查询逻辑。

[x] 高优先级: 实现 Chainlink 获取 Fabric 真实结果后，调用 FISCO BCOS 合约的物理回写机制。

[ ] 提取全网硬编码，重构为基于 Viper 的统一 YAML 配置中心。

[ ] 将各 Golang 微服务节点进行 Alpine 轻量级 Docker 化打包。

[ ] 在适配器调用层加入指数退避重试 (Exponential Backoff) 机制，提升应对网络抖动的容灾能力。
