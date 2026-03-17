# 🗺️ 虚拟机全局目录导航图 (Directory Map)

为了防止在复杂的双链底层环境中迷失，特建立此全局目录索引。本指南标明了所有核心组件在 Ubuntu 虚拟机中的绝对物理路径。

---

## 📍 1. 跨链预言机大本营 (本项目所在)
- **路径**: `~/cross-chain-project/`
- **说明**: 我们的 GitHub 战略基地与 Go 微服务预言机代码所在地。
- **核心文件**:
  - `oracle_main.go`：预言机微服务守护程序（核心代码）。
  - `README.md`：项目说明与排雷血泪史。
  - `ARCHITECTURE.md`：核心架构与演进路线图。
  - `DIRECTORY.md`：当前您正在看的目录地图。

---

## 📍 2. 源链：Hyperledger Fabric 阵地
- **网络主路径**: `~/fabric-project/fabric-samples/test-network/`
- **说明**: 权威存证链的底层网络。每次开机需要在这里执行 `./network.sh up createChannel` 唤醒容器。
- **关键子目录**:
  - `organizations/peerOrganizations/org1.example.com/`：Org1 的 MSP 证书和公私钥所在目录（预言机强依赖此目录的证书去签名查证）。
  - `chaincode/` (或你在部署时指定的路径)：`realcert` (pkicert.go / spbft.go) 智能合约的存放地。

---

## 📍 3. 目标链：FISCO BCOS 阵地
- **节点主路径**: `~/fisco/nodes/127.0.0.1/`
- **说明**: 业务目标链的底层节点集群。每次开机需在此执行 `bash start_all.sh` 唤醒节点。
- **控制台路径**: `~/console/`
- **说明**: Java 交互式控制台，预言机就是通过 `os/exec` 调用这里的 `start.sh` 来实现状态写入的。
- **关键子目录**:
  - `~/console/contracts/solidity/`：`CertOracle.sol` 等业务智能合约的存放和编译目录。

---

## 📍 4. 极客自动化脚本区
- **关机清理脚本**: `~/poweroff_clean.sh`
- **说明**: 强迫症专属的优雅关机脚本，负责拆除 Fabric 容器、超度 FISCO 僵尸进程、并检查 GitHub 是否存盘。强烈建议每次结束战斗时通过此脚本关机。

---

## 💡 极客备忘 (Cheat Sheet)

### 🚀 每日点火连招
```bash
# 1. 斩断时间刺客
sudo timedatectl set-ntp no && sudo date -s "$(curl -sI baidu.com | grep -i '^date:' | cut -d' ' -f2-7)" && sudo timedatectl set-ntp yes

# 2. 唤醒 FISCO
cd ~/fisco/nodes/127.0.0.1/ && bash start_all.sh

# 3. 唤醒 Fabric (如容器已拆除，需重新跑 network.sh up...)
docker start $(docker ps -aq)

# 4. 启动预言机监听
cd ~/cross-chain-project && go run oracle_main.go
