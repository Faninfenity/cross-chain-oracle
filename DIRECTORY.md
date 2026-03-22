# 目录结构说明 (Directory Structure)

本项目包含 FISCO BCOS 到 Hyperledger Fabric 跨链预言机的所有核心组件、适配器及自动化脚本。以下是 `cross-chain-oracle` 仓库的全局文件骨架与模块说明：

```text
cross-chain-oracle/
├── ARCHITECTURE.md        # 系统整体跨链架构设计说明
├── DIRECTORY.md           # 当前目录结构详细说明
├── README.md              # 项目实验日志、待办事项与每日开发记录
├── chainlink-adapter/     # Fabric 外部适配器服务 (运行于宿主机，监听 8081 端口)
│   ├── adapter.go         # 接收 Chainlink Webhook 请求并对接 Fabric SDK 的核心逻辑
│   ├── go.mod
│   └── go.sum
├── chainlink-node/        # Chainlink 预言机节点部署环境
│   └── docker-compose.yml # Chainlink 节点与 PostgreSQL 数据库 Docker 启动配置
├── contracts/             # 跨链预言机通用智能合约目录
├── fisco-contracts/       # FISCO BCOS 专属跨链智能合约目录 (包含 CrossChainOracle.sol)
├── listener/              # FISCO BCOS 自动监听与触发中枢
│   ├── auto_trigger.go    # 轮询监听区块事件，提取 Hash 并自动触发 Chainlink Webhook
│   ├── config.toml        # FISCO BCOS 节点连接配置
│   ├── go.mod
│   └── go.sum
├── scripts/               # 辅助部署与环境测试脚本目录
├── fire.sh                # 跨链请求快捷触发测试脚本 (模拟向 FISCO 发送跨链交易)
├── oracle_main.go         # 预言机主程序入口文件
├── poweroff_clean.sh      # 关机前的环境清理与进程释放脚本
├── start_all.sh           # 全局环境 (Listener, Adapter, 节点) 一键启动脚本
├── start_oracle.sh        # 预言机服务专项启动脚本
├── startup.sh             # 系统环境初始化配置脚本
├── stop_all.sh            # 全局跨链进程与容器一键关停脚本
├── go.mod                 # 根目录 Go 模块依赖声明
└── go.sum                 # 根目录 Go 模块版本校验

