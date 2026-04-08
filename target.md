# 系统扩展计划 · 开发笔记
> 2026-04-08 整理，边做边更新

---

## 现在系统有什么

```
Fabric (L0)      证书存证、吊销、查询
FISCO (L1)       发起查验请求、接收预言机回写结果
Chainlink        主动监听 FISCO 事件，驱动 Fabric 查验
IPFS             物理文件存储，链上只存 CID
```

已经能跑通的功能：
- 存证 / 吊销 / 三态查验（VALID / REVOKED / NOTFOUND）
- auto_trigger 主动轮询 FISCO 区块，捕获 CrossDomainRequested 事件
- config.toml 统一配置
- issuer_ui 证书台账 + verifier_ui 跨链核验

---

## 还缺什么（按优先级）

### 1. DID 身份体系

**问题**：现在 Owner 字段就是个字符串 `"LaoFan_University"`，没法验证这个 CA 的身份是不是真的。

**要做什么**：
- DID 格式定好：`did:pki:org1:Qm8enVq...`
- pkicert.go 加三个接口：
  - `RegisterDID(did, publicKey, orgID)` — 注册
  - `ResolveDID(did)` — 解析，返回公钥等信息
  - `DeactivateDID(did)` — 注销
- Certificate 结构体的 Owner 字段改成 IssuerDID
- 查验时用 DID Document 里的公钥验证证书签名

**工作量**：1~1.5 天

---

### 2. CA 信誉分（放在 FISCO）

**为什么放 FISCO 不放 Fabric**：
- Fabric 是客观事实层，证书在不在、有没有被吊销，一旦写入就是最终结论
- 信誉分是主观评价，会动态变化、需要多方投票，属于业务逻辑，放 FISCO 合适

**新合约 CAReputation.sol**：

```solidity
struct CARecord {
    string  caDID;
    int256  score;           // 0~100，初始100
    uint256 totalIssued;
    uint256 totalRevoked;
    uint256 totalReported;
    string  status;          // TRUSTED / WARNING / BLOCKED
    uint256 lastUpdated;
}
```

**扣分规则**：

| 事件 | 变动 | 触发方式 |
|------|------|---------|
| 旗下证书被吊销 | -5 | 预言机自动 |
| 旗下证书被举报为伪造 | -20 | 投票通过后 |
| 吊销率超过 30% | -10 | 系统定期检查 |
| 监管机构季度审查通过 | +5 | 手动录入 |
| 单日变动上限 | ±10 | 防集中攻击 |

**状态阈值**：
- ≥ 60 → TRUSTED，绿色
- 30~59 → WARNING，黄色
- < 30 → BLOCKED，红色，即使证书本身有效也拒绝

**查验结果从三态升级为五态**：

| 证书状态 | CA 信誉 | 结果 |
|---------|--------|------|
| VALID | ≥60 | 完全信任 ✅ |
| VALID | 30~59 | 谨慎信任，CA 预警 ⚠️ |
| VALID | <30 | 拒绝，CA 不可信 ❌ |
| REVOKED | 任意 | 已吊销 🟠 |
| NOTFOUND | 任意 | 不存在 ❌ |

**工作量**：2~2.5 天

---

### 3. 动态扩展 CA（Fabric 动态添加组织）

**问题**：现在网络里固定只有 Org1 + Org2，新 CA 机构想加入没有机制。

**要做什么**：
- 新 CA 机构准备好 MSP 材料（证书、公钥）
- 现有成员（Org1 + Org2）投票同意
- 用 `configtxlator` 工具更新 channel 配置
- 新组织加入 mychannel，参与背书
- 不需要重启网络

**CA 完整生命周期**：
```
申请加入 → 投票通过 → 动态加入网络
    ↓
注册 DID（身份）
    ↓
初始化信誉分 100
    ↓
日常运行，信誉分动态变化
    ↓
信誉分跌破阈值 → 降级预警
    ↓
Group1 监管层裁决 → 封禁 or 恢复
    ↓
彻底封禁 → 动态移出网络（更新 channel 配置踢出组织）
```

**涉及的 Fabric 操作**：
- `configtxlator` 解码/编码 channel 配置
- `peer channel update` 提交配置变更
- MSP 配置里的 `cacerts/`（白名单）和 `crls/`（黑名单）管理

**工作量**：1.5~2 天

---

### 4. 多群组多共识（FISCO）

**为什么需要换共识**：CA 信誉投票时，参与机构互不信任，需要 PBFT 保证即使有节点被黑也无法操纵投票结果。不是为了换而换，是投票场景的实际需要。

**三层群组划分**：

| 群组 | 共识 | 节点数 | 成员 | 职责 |
|------|------|--------|------|------|
| Group1 监管层 | PBFT | 3~5 | 工信部、行业协会 | 最终裁决封禁/恢复 |
| Group2 运营层 | PBFT | 10~20 | 三大运营商、设备商 | 日常投票扣分 |
| Group3 接入层 | Raft | 100+ | 企业、大学、医院 | 提交举报证据，不能直接扣分 |

**防恶意攻击**：
- 举报必须附问题证书 CID，无证据自动驳回
- 同一机构对同一 CA 30 天内最多举报 3 次
- 恶意举报被驳回 3 次，投票权降级
- 单日最多被扣 10 分
- 分级投票权重：监管×3，运营商×2，普通×1
- PBFT 保证控制 1/3 以下节点没用

**工作量**：2~3 天

---

## 三个核心概念的关系

```
证书  → 解决「你是谁」
CA    → 解决「谁来担保你」
CA信誉 → 解决「担保人靠不靠谱」
```

三个扩展方案的逻辑链：
```
DID（身份）→ CA信誉评估（行为）→ 多共识（投票公正性）
```

---

## 建议执行顺序

```
第一步  DID 规范化 + RegisterDID 链码接口
第二步  CAReputation.sol 合约 + 预言机自动扣分
第三步  verifier_ui 五态结果展示
第四步  Fabric 动态添加组织（configtxlator）
第五步  FISCO 多群组配置（Group1/2/3）
第六步  投票合约（举报、扣分、防攻击）
第七步  签名验证（用 DID 公钥验证证书）
第八步  联调 + 压测
```

---

## 待确认的问题

- [ ] DID 公钥存储格式：PEM 还是 hex？
- [ ] 新 CA 加入网络的背书策略怎么更新（现在是 Org1+Org2 都要签）
- [ ] CAReputation 合约部署地址需要记录到 config.toml
- [ ] Group1/2/3 的初始成员怎么确定（论文里的场景）
- [ ] 信誉分初始化时机：注册 DID 时？还是第一次签发证书时？
