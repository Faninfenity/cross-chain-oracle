```mermaid
sequenceDiagram
    autonumber

    %% 参与节点定义
    participant Issuer as 颁发机构 (企业/高校)
    participant HashEngine as 哈希引擎 / IPFS节点
    participant ClientSDK as Fabric Client SDK
    participant Peer as Fabric 背书节点 (Endorsing Peer)
    participant Orderer as Fabric 排序节点 (Orderer)
    participant Ledger as Fabric 底层账本 (World State)

    Note over Issuer, Ledger: 源头数据铸造阶段：Fabric 权威存证过程

    %% 第一步：文件指纹提取
    rect rgb(240, 248, 255)
    Issuer->>HashEngine: 提交源文件 (毕业证/合同等物理文件)
    HashEngine-->>Issuer: 返回高强度数字指纹 (SHA-256 Hash 或 IPFS CID)
    Note left of Issuer: 原文件物理隔离，仅取数字指纹流转
    end

    %% 第二步：发起交易提案
    rect rgb(240, 255, 240)
    Issuer->>ClientSDK: 触发颁发业务逻辑，传入指纹与元数据
    ClientSDK->>Peer: 调用 IssueCertificate 发送交易提案 (Proposal)
    Peer->>Peer: 运行链码 (pkicert.go) 模拟执行交易，校验颁发者权限
    Peer-->>ClientSDK: 返回带有节点签名的背书响应 (包含读写集)
    end

    %% 第三步：共识排序与打包
    rect rgb(255, 250, 205)
    ClientSDK->>Orderer: 收集足够背书后，将交易提交至排序服务
    Orderer->>Orderer: 验证交易，对全网并发交易进行全局排序
    Orderer->>Orderer: 切割并打包生成新区块 (Block)
    end

    %% 第四步：账本固化与落块
    rect rgb(255, 245, 238)
    Orderer->>Peer: 将新区块广播分发给所有 Peer 节点
    Peer->>Peer: 进行 MVCC 验证，确保读写集无冲突
    Peer->>Ledger: 验证通过，状态永久写入世界状态字典库与区块链底座
    Peer-->>ClientSDK: 触发 Block Event，异步推送落块成功回执
    ClientSDK-->>Issuer: 业务系统提示：文件指纹确权存证完毕
    end
