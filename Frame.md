```mermaid
flowchart TB
    %% 样式定义
    classDef frontend fill:#0f172a,stroke:#38bdf8,stroke-width:2px,color:#fff
    classDef fisco fill:#1e3a8a,stroke:#60a5fa,stroke-width:2px,color:#fff
    classDef middle fill:#4c1d95,stroke:#a78bfa,stroke-width:2px,color:#fff
    classDef chainlink fill:#065f46,stroke:#34d399,stroke-width:2px,color:#fff
    classDef fabric fill:#171717,stroke:#a3e635,stroke-width:2px,color:#fff
    classDef db fill:#991b1b,stroke:#f87171,stroke-width:2px,color:#fff

    subgraph UI ["🖥️ 表现层: Web 可视化大屏"]
        Browser["前端浏览器 (Web UI)"]
        Crypto["Web Crypto API<br/>(纯本地提取 SHA-256)"]
        WebGo["web_ui.go<br/>(总调度后台 Port: 8888)"]
        
        Browser -- "拖拽存证文件" --> Crypto
        Crypto -- "GET /api/trigger?id=hash" --> WebGo
        Browser -- "GET /api/query" --> WebGo
    end

    subgraph FISCO_Zone ["⛓️ 业务侧链: FISCO BCOS"]
        Console["Java 控制台<br/>(console.sh)"]
        Contract["CrossChainClient.sol<br/>(核心智能合约)"]
        EventDB[("区块事件日志<br/>(Event Logs)")]
        StateDB[("链上状态账本<br/>(verifyResults)")]

        WebGo -- "os/exec (请求发车/查证)" --> Console
        Console -- "requestVerification" --> Contract
        Contract -- "抛出跨链需求" --> EventDB
        Contract -- "读/写" --> StateDB
    end

    subgraph Middleware ["🚀 跨链中枢: Go 原生守护进程阵列"]
        direction TB
        Listener["监听哨兵 (auto_trigger.go)"]
        Writer["回写中枢 (fisco_writer.go | Port: 8082)"]

        subgraph Filter ["🛡️ 核心防御引擎"]
            LenFilter{"Input Data 长度<br/>> 230 ?"}
            ABIDecoder["ABI Hex 极客解码器<br/>(剥离动态指纹)"]
            AntiLoop(("拦截套娃<br/>(丢弃)"))
        end

        Listener -- "1秒/次 轮询区块" --> EventDB
        Listener --> LenFilter
        LenFilter -- "Yes (系回写交易)" --> AntiLoop
        LenFilter -- "No (系合法跨链)" --> ABIDecoder
    end

    subgraph Chainlink_Zone ["🔮 路由层: Chainlink 预言机"]
        Webhook["Webhook Initiator<br/>(API Port: 6688)"]
        Pipeline["TOML 任务流水线<br/>(JSON Parse -> Bridge)"]
        
        ABIDecoder -- "携带 Session POST" --> Webhook
        Webhook --> Pipeline
    end

    subgraph Fabric_Zone ["⛓️ 权威存证链: Hyperledger Fabric"]
        Adapter["宿主机穿透适配器<br/>(adapter.go | Port: 8081)"]
        Peer["Fabric Peer 节点<br/>(localhost:7051)"]
        Chaincode["pki 智能合约<br/>(QueryCertificate)"]
        FabricDB[("Fabric World State<br/>(终极确权账本)")]

        Pipeline -- "HTTP 路由转发目标 Hash" --> Adapter
        Adapter -- "os/exec 直调 peer query<br/>(绕过 SDK 限制)" --> Peer
        Peer --> Chaincode
        Chaincode --> FabricDB
    end

    %% 回写与闭环链路
    FabricDB -. "返回判决 (True/False)" .-> Adapter
    Adapter -- "组装 JSON (携单号)" --> Pipeline
    Pipeline -- "携带防弹衣 Token POST" --> Writer
    Writer -- "严格参数隔离 (防 OS 注入)" --> Console
    Console -- "fulfillVerification" --> Contract
