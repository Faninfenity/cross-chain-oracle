````mermaid
graph TD
    %% 定义样式，保持与原图色彩对齐
    classDef UserFill fill:#F0F8FF,stroke:#333,stroke-width:1px,color:black;
    classDef IPFSFill fill:#FFF5EE,stroke:#333,stroke-width:1px,color:black;
    classDef FabricFill fill:#F0FFF0,stroke:#333,stroke-width:1px,color:black;
    classDef OracleFill fill:#FFFACD,stroke:#333,stroke-width:1px,color:black;
    classDef FISCOFill fill:#E6E6FA,stroke:#333,stroke-width:1px,color:black;

    %% 🏛️ 顶级节点
    System(🏛️ 異構雙鏈跨鏈存證系統 - 全景功能映射图 V1.0)
    System --> Roles
    System --> Stages

    %% 🎭 阶段一：角色定义与角色功能
    subgraph Roles [🎭 组件角色与硬核功能定义]
        UserRole(企業/高校 前端)
        IPFSRole(IPFS 物理存储层)
        FabricRole(Fabric 后端权威大本营)
        OracleRole(Oracle 跨链枢纽中间件)
        FISCORole(FISCO BCOS 前端高频业务链)

        %% 应用样式
        class UserRole UserFill;
        class IPFSRole IPFSFill;
        class FabricRole FabricFill;
        class OracleRole OracleFill;
        class FISCORole FISCOFill;

        %% 详细功能描述
        UserRole -->|职责| UFunc1[源头存证入口 8889]
        UserRole -->|职责| UFunc2[跨域查验大屏 8888]

        IPFSRole -->|职责| IFunc1[星际文件系统分布式存储]
        IPFSRole -->|职责| IFunc2[物理文件CID特征值计算]

        FabricRole -->|职责| FFunc1[权威链码 pkicert.go]
        FabricRole -->|职责| FFunc2[数据权威固化]
        FabricRole -->|职责| FFunc3[S-PBFT 节点信誉判定]

        OracleRole -->|职责| OFunc1[事件捕获 auto_trigger]
        OracleRole -->|职责| OFunc2[跨网提取/状态逻辑研判]
        OracleRole -->|职责| OFunc3[结果签名回写 fisco_writer]

        FISCORole -->|职责| FiFunc1[业务合约 CertOracle.sol]
        FISCORole -->|职责| FiFunc2[高频查验缓冲]
        FISCORole -->|职责| FiFunc3[本地历史共识固化]
    end

    %% 🔄 阶段二：详细数据流转与功能咬合
    subgraph Stages [🔄 全景数据流转 5 步曲之功能咬合]
        direction TB

        %% 阶段一：源头颁发与存证
        subgraph Stage1 [1️⃣ 源头颁发与存证 Data Induction]
            direction LR
            S1_UI[1.a 前端选择文件上传 issuer_ui.go]
            S1_IPFS[1.b CID计算与存储 api/upload]
            S1_UI2[1.c 返回CID Qm...123]
            S1_Fabric[1.d 链码铸造 pkicert.go -> IssueCertificate]
            
            S1_UI ==>|PDF| S1_IPFS
            S1_IPFS ==>|Qm...123| S1_UI2
            S1_UI2 ==>|调用原生脚本| S1_Fabric
            class Stage1 UserFill;
        end

        %% 阶段二：用户发起跨链查验
        subgraph Stage2 [2️⃣ 用户发起跨链查验 Verification Request]
            direction LR
            S2_User[2.a 网页输入Qm...123发验证 verifier_ui.go]
            S2_FISCO[2.b 合约检索本地 CertOracle.sol]
            S2_Status[2.c 查无此证 (False, False)]
            S2_Event[2.d 抛出事件 CertVerificationRequested]

            S2_User ==>|验证请求| S2_FISCO
            S2_FISCO ==>|本地状态| S2_Status
            S2_Status ==>|触发| S2_Event
            class Stage2 FISCOFill;
        end

        %% 阶段三：预言机捕获与路由
        subgraph Stage3 [3️⃣ 预言机捕获与路由 Middleware Listen]
            direction LR
            S3_Oracle[3.a 事件总线精确监听 auto_trigger.go]
            S3_Parser[3.b 解析Qm...123]
            S3_Router[3.c 跨网直连 Fabric gRPC 节点]

            S3_Oracle ==>|捕获| S3_Parser
            S3_Parser ==>|路由| S3_Router
            class Stage3 OracleFill;
        end

        %% 阶段四：权威核实（状态研判）
        subgraph Stage4 [4️⃣ 权威核实 Consensus Verdict]
            direction LR
            S4_Fabric[4.a 调用 QueryCertificate]
            S4_Status[4.b 返回证书状态与頒發者信誉]
            S4_Verdict{4.c 预言机逻辑研判}
            S4_A(A 判定 True<br/>存在+有效+信誉正常)
            S4_B(B 判定 False<br/>不存在或吊销)

            S4_Fabric ==>|资产数据| S4_Status
            S4_Status ==>|评估| S4_Verdict
            S4_Verdict -->|合格| S4_A
            S4_Verdict -->|非法| S4_B
            class Stage4 FabricFill;
        end

        %% 阶段五：结果回写与完美闭环
        subgraph Stage5 [5️⃣ 结果回写与完美闭环 Consensus Lock]
            direction LR
            S5_Oracle[5.a 私钥签名回写 verifyCert fisco_writer.go]
            S5_FISCO[5.b 状态同步完成落块 CertOracle.sol]
            S5_Front[5.c 前端展示验证结果]

            S5_Oracle ==>| True/False | S5_FISCO
            S5_FISCO ==>|状态监听| S5_Front
            class Stage5 OracleFill;
        end

        %% 串联五个阶段
        Stage1 ==> Stage2
        Stage2 ==> Stage3
        Stage3 ==> Stage4
        Stage4 ==> Stage5
    end
````
