#!/bin/bash

# ====================================================
# 5G-A 异构跨域信任底座 - 全域日志聚合监控探针 v1.0
# ====================================================

# 定义颜色输出规范
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # 恢复默认颜色

clear
echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}      [5G-A 跨域总线] 神经节点日志监控探针就绪      ${NC}"
echo -e "${CYAN}======================================================${NC}"
echo -e "请选择你要监听的系统神经元节点："
echo -e "${GREEN}  1) [上帝视角] 全局瀑布流监控 (同时监听大屏、传达室、回写中枢)${NC}"
echo -e "${YELLOW}  2) [传达室] 仅监听 Auto Trigger (事件拦截与预言机触发)${NC}"
echo -e "${YELLOW}  3) [回写中枢] 仅监听 FISCO Writer (双阶段落块执行)${NC}"
echo -e "${YELLOW}  4) [查证大屏] 仅监听 Verifier UI (前端跨域请求与响应)${NC}"
echo -e "  5) 退出监控"
echo -e "${CYAN}======================================================${NC}"

read -p "请输入对应的战术编号 [1-5]: " choice

echo -e "${CYAN}[系统] 正在建立异构网络物理连接，按 Ctrl+C 可随时切断监听并返回主控台...${NC}\n"

case $choice in
    1)
        # 使用 tail -f 同时追踪三个核心微服务的日志，Bash 会自动用文件名作为段落前缀
        tail -f logs/verifier_ui.log logs/auto_trigger.log logs/fisco_writer.log
        ;;
    2)
        tail -f logs/auto_trigger.log
        ;;
    3)
        tail -f logs/fisco_writer.log
        ;;
    4)
        tail -f logs/verifier_ui.log
        ;;
    5)
        echo -e "${GREEN}[退出] 监控探针已安全脱离。${NC}"
        exit 0
        ;;
    *)
        echo -e "${YELLOW}[警告] 无效的战术指令，连接中止。${NC}"
        exit 1
        ;;
esac
