#!/bin/bash
# BotMatrix Docker 容器管理脚本
# 提供简化的容器操作命令

set -e

# 容器名称映射表
declare -A CONTAINER_MAP=(
    ["manager"]="btmgr"
    ["system"]="btsys"
    ["qq"]="btqq"
    ["wechat"]="btwc"
    ["wechat-go"]="btwg"
    ["dingtalk"]="btdt"
    ["feishu"]="btfs"
    ["telegram"]="bttg"
    ["discord"]="btdc"
    ["slack"]="btsl"
    ["kook"]="btkk"
    ["email"]="btem"
    ["wecom"]="btwm"
)

# 反向映射（容器名到服务名）
declare -A REVERSE_MAP=(
    ["btmgr"]="manager"
    ["btsys"]="system"
    ["btqq"]="qq"
    ["btwc"]="wechat"
    ["btwg"]="wechat-go"
    ["btdt"]="dingtalk"
    ["btfs"]="feishu"
    ["bttg"]="telegram"
    ["btdc"]="discord"
    ["btsl"]="slack"
    ["btkk"]="kook"
    ["btem"]="email"
    ["btwm"]="wecom"
)

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

show_help() {
    cat << EOF
BotMatrix Docker 管理脚本

用法: $0 <操作> [服务名] [--all]

操作:
  start    - 启动容器
  stop     - 停止容器  
  restart  - 重启容器
  status   - 查看状态（默认）
  logs     - 查看日志
  shell    - 进入容器shell
  up       - 启动所有服务
  down     - 停止所有服务
  clean    - 清理所有容器和镜像

服务名:
  manager     - 管理平台 (btmgr)
  system      - 系统服务 (btsys)
  qq          - QQ机器人 (btqq)
  wechat      - 微信机器人 (btwc)
  wechat-go   - 微信Go版 (btwg)
  dingtalk    - 钉钉机器人 (btdt)
  feishu      - 飞书机器人 (btfs)
  telegram    - Telegram机器人 (bttg)
  discord     - Discord机器人 (btdc)
  slack       - Slack机器人 (btsl)
  kook        - Kook机器人 (btkk)
  email       - 邮件机器人 (btem)
  wecom       - 企业微信机器人 (btwm)

示例:
  $0 status                    # 查看所有容器状态
  $0 restart manager            # 重启管理平台
  $0 logs wechat              # 查看微信机器人日志
  $0 shell manager             # 进入管理平台shell
  $0 up --all                   # 启动所有服务
  $0 stop qq                    # 停止QQ机器人

快捷命令:
  docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  docker restart <容器名>
  docker logs -f <容器名>
EOF
}

get_container_status() {
    echo -e "${GREEN}容器状态一览:${NC}"
    echo -e "${GREEN}=============${NC}"
    
    # 获取所有相关容器
    local containers=$(docker ps -a --format "{{.Names}}" | grep -E "^(btmgr|btsys|btqq|btwc|btwg|btdt|btfs|bttg|btdc|btsl|btkk|btem|btwm)$" || true)
    
    if [ -z "$containers" ]; then
        echo -e "${YELLOW}未找到BotMatrix容器${NC}"
        return
    fi
    
    while IFS= read -r container; do
        if [ -n "$container" ]; then
            local service_name="${REVERSE_MAP[$container]:-unknown}"
            local status=$(docker inspect -f '{{.State.Status}}' "$container" 2>/dev/null || echo "unknown")
            local health=$(docker inspect -f '{{.State.Health.Status}}' "$container" 2>/dev/null || echo "")
            local ports=$(docker inspect -f '{{range $p, $conf := .NetworkSettings.Ports}}{{$p}} -> {{(index $conf 0).HostPort}} {{end}}' "$container" 2>/dev/null || echo "")
            
            local status_color="${GREEN}"
            if [ "$status" != "running" ]; then
                status_color="${RED}"
            fi
            
            printf "%-12s (%s): %s\n" "$service_name" "$container" "$status" | sed "s/$status/$status_color$status${NC}/"
            
            if [ -n "$health" ] && [ "$health" != "" ]; then
                local health_color="${GREEN}"
                if [ "$health" != "healthy" ]; then
                    health_color="${RED}"
                fi
                echo -e "  健康: ${health_color}$health${NC}"
            fi
            
            if [ -n "$ports" ] && [ "$ports" != "" ]; then
                echo -e "  端口: ${GRAY}$ports${NC}"
            fi
        fi
    done <<< "$containers"
}

start_container() {
    local container=$1
    echo -e "${GREEN}启动容器: $container${NC}"
    docker start "$container"
}

stop_container() {
    local container=$1
    echo -e "${YELLOW}停止容器: $container${NC}"
    docker stop "$container"
}

restart_container() {
    local container=$1
    echo -e "${CYAN}重启容器: $container${NC}"
    docker restart "$container"
}

show_logs() {
    local container=$1
    echo -e "${GREEN}查看日志: $container (按 Ctrl+C 退出)${NC}"
    docker logs -f --tail 100 "$container"
}

enter_shell() {
    local container=$1
    echo -e "${GREEN}进入容器shell: $container${NC}"
    docker exec -it "$container" /bin/sh
}

start_all_services() {
    echo -e "${GREEN}启动所有BotMatrix服务...${NC}"
    docker-compose up -d
}

stop_all_services() {
    echo -e "${YELLOW}停止所有BotMatrix服务...${NC}"
    docker-compose down
}

clean_all() {
    echo -e "${RED}清理所有BotMatrix容器和镜像...${NC}"
    echo -e "${RED}警告: 这将删除所有相关容器和镜像！${NC}"
    read -p "确认清理? (输入 'yes' 确认): " confirm
    if [ "$confirm" = "yes" ]; then
        docker-compose down --rmi all --volumes
        echo -e "${GREEN}清理完成${NC}"
    else
        echo -e "${YELLOW}操作已取消${NC}"
    fi
}

# 主逻辑
ACTION=${1:-status}
SERVICE=${2:-}
ALL=false

# 处理参数
if [ "$SERVICE" = "--all" ]; then
    ALL=true
    SERVICE=""
elif [ "$#" -gt 2 ] && [ "${3:-}" = "--all" ]; then
    ALL=true
fi

if [ "$ACTION" = "help" ] || [ "$ACTION" = "-h" ] || [ "$ACTION" = "--help" ]; then
    show_help
    exit 0
fi

# 处理服务名到容器名的转换
CONTAINER_NAME=""
if [ -n "$SERVICE" ]; then
    if [ "${CONTAINER_MAP[$SERVICE]:-}" != "" ]; then
        CONTAINER_NAME="${CONTAINER_MAP[$SERVICE]}"
    else
        # 可能是直接的容器名
        CONTAINER_NAME="$SERVICE"
    fi
fi

# 执行操作
case "$ACTION" in
    'status')
        if [ -n "$SERVICE" ]; then
            # 检查特定容器状态
            local status=$(docker inspect -f '{{.State.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "unknown")
            if [ "$status" != "unknown" ]; then
                echo -e "$CONTAINER_NAME 状态: $status"
            else
                echo -e "${RED}容器 $CONTAINER_NAME 不存在${NC}"
            fi
        else
            get_container_status
        fi
        ;;
    'start')
        if [ "$ALL" = true ]; then
            start_all_services
        elif [ -n "$SERVICE" ]; then
            start_container "$CONTAINER_NAME"
        else
            echo -e "${RED}请指定服务名或使用 --all 参数${NC}"
            exit 1
        fi
        ;;
    'stop')
        if [ "$ALL" = true ]; then
            stop_all_services
        elif [ -n "$SERVICE" ]; then
            stop_container "$CONTAINER_NAME"
        else
            echo -e "${RED}请指定服务名或使用 --all 参数${NC}"
            exit 1
        fi
        ;;
    'restart')
        if [ -n "$SERVICE" ]; then
            restart_container "$CONTAINER_NAME"
        else
            echo -e "${RED}请指定服务名${NC}"
            exit 1
        fi
        ;;
    'logs')
        if [ -n "$SERVICE" ]; then
            show_logs "$CONTAINER_NAME"
        else
            echo -e "${RED}请指定服务名${NC}"
            exit 1
        fi
        ;;
    'shell')
        if [ -n "$SERVICE" ]; then
            enter_shell "$CONTAINER_NAME"
        else
            echo -e "${RED}请指定服务名${NC}"
            exit 1
        fi
        ;;
    'up')
        start_all_services
        ;;
    'down')
        stop_all_services
        ;;
    'clean')
        clean_all
        ;;
    *)
        show_help
        ;;
esac