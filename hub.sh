#!/bin/bash

# Docker Hub部署脚本
# Git代码文件代理服务 - 使用Docker Hub镜像

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

IMAGE_NAME="vansour/ghproxy"
COMPOSE_FILE="docker-compose.hub.yml"

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose 未安装，请先安装Docker Compose"
        exit 1
    fi
}

pull_image() {
    print_status "从Docker Hub拉取最新镜像..."
    docker pull $IMAGE_NAME:latest
    if [ $? -eq 0 ]; then
        print_success "镜像拉取成功"
    else
        print_error "镜像拉取失败"
        exit 1
    fi
}

start_service() {
    print_status "启动服务..."
    
    # 创建日志目录
    mkdir -p ./logs
    
    # 启动服务
    docker-compose -f $COMPOSE_FILE up -d
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    if docker-compose -f $COMPOSE_FILE ps | grep -q "Up"; then
        print_success "服务启动成功！"
        print_success "服务地址: http://127.0.0.1:8080"
        print_success "Web界面: http://服务器IP:8080"
        print_status "查看日志: ./hub.sh logs"
    else
        print_error "服务启动失败"
        docker-compose -f $COMPOSE_FILE logs
        exit 1
    fi
}

stop_service() {
    print_status "停止服务..."
    docker-compose -f $COMPOSE_FILE down
    print_success "服务已停止"
}

restart_service() {
    print_status "重启服务..."
    docker-compose -f $COMPOSE_FILE restart
    print_success "服务已重启"
}

show_status() {
    print_status "服务状态:"
    docker-compose -f $COMPOSE_FILE ps
    echo ""
    
    # 显示资源使用情况
    print_status "资源使用情况:"
    docker stats --no-stream $(docker-compose -f $COMPOSE_FILE ps -q) 2>/dev/null || echo "暂无运行中的容器"
}

show_logs() {
    print_status "显示服务日志 (按Ctrl+C退出):"
    docker-compose -f $COMPOSE_FILE logs -f
}

update_service() {
    print_status "更新服务..."
    docker-compose -f $COMPOSE_FILE down
    pull_image
    docker-compose -f $COMPOSE_FILE up -d
    print_success "服务更新完成"
}

cleanup() {
    print_status "清理Docker资源..."
    docker-compose -f $COMPOSE_FILE down -v --remove-orphans
    docker system prune -f
    print_success "清理完成"
}

quick_deploy() {
    print_status "快速部署..."
    check_docker
    pull_image
    start_service
    print_success "快速部署完成！"
}

show_help() {
    echo "Git代码文件代理服务 - Docker Hub部署脚本"
    echo ""
    echo "用法: $0 {deploy|pull|start|stop|restart|status|logs|update|cleanup|help}"
    echo ""
    echo "命令说明:"
    echo "  deploy    - 快速部署（拉取镜像并启动）"
    echo "  pull      - 拉取最新镜像"
    echo "  start     - 启动服务"
    echo "  stop      - 停止服务"
    echo "  restart   - 重启服务"
    echo "  status    - 查看服务状态"
    echo "  logs      - 查看服务日志"
    echo "  update    - 更新服务"
    echo "  cleanup   - 清理Docker资源"
    echo "  help      - 显示帮助信息"
    echo ""
    echo "镜像信息:"
    echo "  Docker Hub: https://hub.docker.com/r/vansour/ghproxy"
    echo "  镜像名称: $IMAGE_NAME"
    echo "  端口: 8080"
    echo "  支持平台: GitHub, GitLab, Hugging Face, SourceForge"
    echo ""
    echo "快速开始:"
    echo "  curl -O https://raw.githubusercontent.com/your-repo/ghproxy/main/hub.sh"
    echo "  chmod +x hub.sh"
    echo "  ./hub.sh deploy"
}

# 检查Docker环境
check_docker

# 处理参数
case "$1" in
    deploy)
        quick_deploy
        ;;
    pull)
        pull_image
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    update)
        update_service
        ;;
    cleanup)
        cleanup
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "用法: $0 {deploy|pull|start|stop|restart|status|logs|update|cleanup|help}"
        echo "使用 '$0 help' 查看详细帮助"
        exit 1
        ;;
esac
