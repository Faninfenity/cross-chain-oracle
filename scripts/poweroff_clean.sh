#!/bin/bash

echo "====================================================="
echo "🛑 开始执行【优雅关机】程序... 正在清理战场"
echo "====================================================="

# 1. 撤收 Fabric 阵地
echo "📦 [1/3] 正在拆除 Fabric 网络与 Docker 容器..."
if [ -d "~/fabric-project/fabric-samples/test-network" ]; then
    cd ~/fabric-project/fabric-samples/test-network
    ./network.sh down > /dev/null 2>&1
    docker volume prune -f > /dev/null 2>&1
    echo "✅ Fabric 已安全撤收。"
else
    echo "⚠️ 未找到 Fabric 目录，跳过。"
fi

# 2. 物理超度 FISCO 僵尸
echo "🏢 [2/3] 正在停止 FISCO BCOS 节点并清理残留进程..."
pkill -9 fisco-bcos > /dev/null 2>&1
echo "✅ FISCO 进程已物理超度。"

# 3. GitHub 存盘检查
echo "☁️ [3/3] 正在检查 GitHub 存盘情况..."
cd ~/cross-chain-project
if [[ -n $(git status -s) ]]; then
    echo "❌ 警告：你还有未提交的代码改动！"
    echo "请先执行 git add/commit/push，或者输入 'c' 强行关机，其他键退出脚本。"
    read -n 1 user_choice
    if [[ "$user_choice" != "c" ]]; then
        echo -e "\n🛑 关机程序已取消，先去推代码吧！"
        exit 1
    fi
else
    echo "✅ 代码已全部存盘。"
fi

echo -e "\n🚀 场清理完毕！虚拟机即将在 3 秒后断电..."
sleep 3
sudo poweroff
