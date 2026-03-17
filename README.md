# 🚀 异构双链跨链存证系统 MVP (Fabric <-> FISCO BCOS)

## 📖 项目简介
本项目是一个基于 Go 语言编写的轻量级跨链预言机（Oracle），旨在实现 Hyperledger Fabric（权威存证链）与 FISCO BCOS（前端业务链）之间的数据互通与信任传递。

## 🏆 极客排雷四大天坑记录 (The Debugging Epic)
1. **x509 证书时差刺客**：快照恢复导致系统时间错乱。（解法：强行同步 `timedatectl`）。
2. **20200 端口僵尸节点**：回滚快照后，老节点驻留内存。（解法：`pkill -9 fisco-bcos`）。
3. **Legacy SDK 黑盒陷阱**：旧版 Fabric SDK 无法解析 v2.5 的 `config.yaml`。（解法：拥抱 `Fabric Gateway`，动态盲抓公私钥）。
4. **FISCO 控制台暗改参数**：v2.9.2 版本控制台剥夺了 `-f` 剧本参数。（解法：使用 Linux 管道符 `|` 与 `echo -e` 强行塞入交互指令）。

## ⚙️ 核心架构与代码
- 源链：Hyperledger Fabric v2.5.4
- 目标链：FISCO BCOS v2.9
- 预言机机制：Go 语言直连 Fabric 获取状态 -> JSON 解析 -> 唤醒本地 Java 控制台写入 FISCO。

## 🚀 运行方式
```bash
go run oracle_main.go
