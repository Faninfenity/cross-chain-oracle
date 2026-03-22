# Cross-Chain Oracle 实验与开发日志

**项目名称**: cross-chain-oracle
**开发者**: Faninfenity
**核心技术栈**: Go, C++, Solidity, Docker, Hyperledger Fabric, FISCO BCOS, Chainlink

## 1. 系统架构与当前状态 (Context Snapshot)
本项目旨在实现 FISCO BCOS 与 Hyperledger Fabric 之间的跨链数据核查。当前架构由以下四个核心组件构成：

* **FISCO BCOS (业务发起端)**:
  * 部署了预言机合约 `CrossChainOracle.sol`。
  * 当前合约地址: `0xb59e0050aC3449D8A9F7A40670ed86DA7D89d5Ac`。
* **监听与触发中枢 (Listener - `auto_trigger.go`)**:
  * 通过 Go SDK 连接 FISCO BCOS 节点。
  * 轮询监听最新区块，拦截发往预言机合约的 `requestVerify` 事件。
  * 提取交易 Hash，通过 HTTP POST 自动触发 Chainlink Webhook。
* **跨链路由中枢 (Chainlink Node)**:
  * 运行在 Docker 容器中，暴露 6688 端口。
  * 账号凭证: `admin@crosschain.local` / `Admin@Chainlink2026`。
  * 核心 Job (Webhook V3): `Fabric-CrossChain-Webhook-V3` (External Job ID: `2352c9c6-0e1c-4b2e-b867-c86cbcb96820`)。
  * 配置了名为 `fabric-adapter` 的 Bridge。
* **Fabric 外部适配器 (Adapter - `adapter.go`)**:
  * 运行在宿主机 8081 端口 (`http://host.docker.internal:8081/`)。
  * 接收来自 Chainlink Bridge 的核查请求，目前使用 `mockFabricVerify` 模拟返回 `isValid: true` 结果。

## 2. 待办事项 (TODO List)
* [ ] **高优先级**: 解决 Chainlink Webhook 调用的 401 Basic Auth 认证拦截问题（排查节点数据库凭证缓存或配置读取机制）。
* [ ] 在 `adapter.go` 中剥离 Mock 逻辑，接入真实的 Fabric Go SDK 账本查询逻辑。
* [ ] 实现 Chainlink 获取 Fabric 结果后，调用 FISCO BCOS 智能合约的回写机制（将 `isValid` 状态上链）。
* [ ] 将配置文件中的明文账密和私钥剥离，完善 `.gitignore` 配置。

---

## 3. 实验日志 (Daily Log)

### [2026-03-22] 自动监听中枢开发与全链路联调攻坚

**核心进展**:
1. **外部适配器 (Adapter) 连通测试**: 成功在 8081 端口启动 `adapter.go`。通过 Chainlink Web UI 手动触发 Webhook，成功穿透 Docker 网络，终端打印 `[Adapter] 收到 Chainlink 跨链核查任务!`。确认 Chainlink 前端 `Failed to parse task graph` 为纯 UI 渲染 Bug，不影响底层逻辑。
2. **监听器 (Listener) 开发**: 编写了 `auto_trigger.go`，实现了对 FISCO BCOS 区块的 7x24 小时轮询监听。
3. **数据解析修复**: 在处理 FISCO Go SDK 返回的 `Block` 对象时，解决了 `tx.Hash` 字段无法直接读取的问题，通过类型断言 `tx, ok := txInterface.(map[string]interface{})` 成功提取出交易 Hash。
4. **代码版本控制**: 将 `chainlink-adapter` 和 `listener` 模块的代码统一汇总，解决了 Git 暂存区冲突 (`fetch first` 与 `rebase` 问题)，成功推送到远程 `cross-chain-oracle` 仓库的主分支。

**技术卡点 (Blocker)**:
* **Webhook 认证穿透失败**: `auto_trigger.go` 在向 Chainlink 发送自动化触发指令时，遭遇 `401 Unauthorized`。
* 尝试在 Go 代码中注入 `req.SetBasicAuth(CL_USER, CL_PASS)`，并使用 `curl -u` 命令行进行裸测，均被拒绝。
* 疑似 Chainlink 节点在首次启动后，将初始凭证固化在 PostgreSQL 数据库卷中，导致外部 API 调用时出现鉴权信息不同步。后续需考虑执行 `docker-compose down -v` 清理数据卷并重建 Job。

### [2026-03-13 至 2026-03-20] 跨链基础设施搭建与合约部署

**核心进展**:
1. 确立了跨链工程的基本拓扑结构，初始化 GitHub 仓库 `git@github.com:Faninfenity/cross-chain-oracle.git`。
2. 梳理了 Hyperledger Fabric, FISCO BCOS, 以及 Chainlink 预言机之间的交互边界。
3. 编写了 FISCO BCOS 端的智能合约接口，并完成了初步的编译与部署验证。
4. 部署并启动了 Chainlink 节点容器环境。
