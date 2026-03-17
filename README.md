# 🚀 异构双链跨链存证系统 (Hyperledger Fabric <-> FISCO BCOS)

![License](https://img.shields.io/badge/license-MIT-blue)
![Version](https://img.shields.io/badge/version-v1.0.0--MVP-green)
![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)

## 📖 项目全景 (Project Overview)
本项目是一个基于 Go 语言研发的**轻量级微服务跨链预言机（Cross-Chain Oracle）**。
旨在打破区块链之间的数据孤岛，实现 **Hyperledger Fabric（作为权威存证源链）** 与 **FISCO BCOS（作为前端业务目标链）** 之间的数据互通、信任传递与状态的原子级搬运。

---

## 🏗️ 系统架构图 (Architecture)

```text
[ 前端/业务系统 ] 
       │ (HTTP GET: /api/verify?hash=Qm...)
       ▼
┌─────────────────────────────────────────────────────┐
│ 🔮 Cross-Chain Oracle (Go 微服务守护进程)           │
│  ├─ 1. HTTP 监听器 (端口: 8080)                     │
│  ├─ 2. Fabric Gateway 引擎 (gRPC 常驻内存连接)      │
│  └─ 3. FISCO BCOS 胶水驱动 (os/exec 进程唤醒)       │
└──────┬───────────────────────────────────────┬──────┘
       │ (1. 权威查证)                         │ (2. 状态写入)
       ▼                                       ▼
【 源链: Hyperledger Fabric 】         【 目标链: FISCO BCOS 】
   - 版本: v2.5.4                        - 版本: v2.9.2
   - 角色: 权威存证/数据锚点             - 角色: 业务流转/状态同步
   - 合约: realcert (Go)                 - 合约: CertOracle (Solidity)
✨ 已经实现的核心功能 (Features Implemented)
gRPC 极速查证：完全抛弃老旧的 config.yaml 黑盒，拥抱新一代 Fabric Gateway SDK，通过动态盲抓公私钥，实现预言机启动即与 Fabric 底层建立高并发长连接。

RESTful API 守护模式：预言机以 7x24h 微服务形态运行，提供标准的 http://localhost:8080/api/verify 接口，支持外部系统随时触发跨链任务。

恶意跨链拦截：具备强大的防御机制，当收到伪造或不存在的哈希验证请求时，预言机会精准拦截并拒绝向 FISCO BCOS 发起写入，保护业务链数据纯洁性。

管道直入式跨链写入：使用极其稳定的 Linux 系统级胶水代码，将解析后的可信状态，通过管道符无缝打入 FISCO BCOS Java 控制台，完成跨链数据上链。

🏆 极客排雷血泪史 (The Debugging Epic)
在搭建这套底层双链并打通预言机的过程中，我们踩平了足以劝退 99% 开发者的底层大坑：

🕒 x509 证书时差刺客：虚拟机快照恢复导致系统时间错乱，Fabric 节点颁发的证书被误判“来自未来”。（破局：引入强行同步 timedatectl 的开机自检机制）。

🧟 20200 端口僵尸节点：回滚快照后，旧版本 FISCO 节点驻留内存，导致新网络端口被霸占。（破局：建立 pkill -9 物理超度与清理脚本）。

📦 Legacy SDK 黑盒陷阱：旧版 Fabric SDK-Go 无法正确解析 v2.5 的底层证书结构。（破局：全面拥抱 Fabric Gateway 降维打击）。

🥷 FISCO 控制台暗改参数：v2.9.2 版本控制台剥夺了 -f 剧本执行参数的含义。（破局：放弃传参，使用 echo -e 结合管道符 |，将命令和 quit 强行塞入控制台喉咙）。

🎯 接下来的演进目标 (Future Roadmap)
虽然 MVP 已经完美闭环，但要走向真正的工业级生产环境，我们的征途才刚刚开始：

[ ] 阶段一：告别控制台 (SDK Native Integration)

目标：废弃当前的 os/exec 调用 Java 控制台的方式。

方案：引入 FISCO BCOS Go SDK，在 Go 语言内存层面直接编译 ABI，实现预言机对两条链的纯原生、毫秒级直接交互。

[ ] 阶段二：智能事件监听 (Event-Driven Mode)

目标：预言机从“被动 API 触发”升级为“主动监听链上事件”。

方案：监听 FISCO BCOS 上的 VerifyRequest 事件，一有业务合约发起请求，预言机自动捕获、自动查证、自动回写，彻底实现无人值守。

[ ] 阶段三：可信执行记录 (Oracle Logging & DB)

目标：记录每一次跨链搬运的历史。

方案：外挂 MySQL/Redis，记录跨链请求的成功率、时间戳和 txHash，用于数据溯源。

[ ] 阶段四：跨链可视化看板 (Dashboard)

目标：给这套硬核的底层系统穿上漂亮的外衣。

方案：编写一个 Vue/React 前端面板，实时展示双链状态和预言机心跳。

Built with passion and persistence by Architect Fan & AI. 2026.
