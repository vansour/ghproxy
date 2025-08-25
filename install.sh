#!/bin/bash

# Git代码文件代理服务安装脚本
# 支持 GitHub、GitLab、Hugging Face、SourceForge

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_NAME="ghproxy"
BINARY_NAME="ghproxy"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
WORK_DIR="/opt/ghproxy"
LOG_DIR="/var/log/ghproxy"

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

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "此脚本需要root权限运行"
        exit 1
    fi
}

install_service() {
    print_status "开始安装 Git代码文件代理服务..."
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装Go语言环境"
        exit 1
    fi
    
    # 编译程序
    print_status "编译程序..."
    go build -o $BINARY_NAME main.go
    if [ $? -ne 0 ]; then
        print_error "编译失败"
        exit 1
    fi
    
    # 创建工作目录
    print_status "创建工作目录..."
    mkdir -p $WORK_DIR
    mkdir -p $LOG_DIR
    
    # 复制可执行文件
    print_status "安装可执行文件到 $INSTALL_DIR..."
    cp $BINARY_NAME $INSTALL_DIR/
    chmod +x $INSTALL_DIR/$BINARY_NAME
    
    # 创建systemd服务文件
    print_status "创建systemd服务..."
    cat > $SERVICE_DIR/$SERVICE_NAME.service << EOF
[Unit]
Description=Git代码文件代理服务
Documentation=https://github.com/
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
WorkingDirectory=$WORK_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$LOG_DIR

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    print_status "重新加载systemd配置..."
    systemctl daemon-reload
    
    # 启用服务
    print_status "启用服务..."
    systemctl enable $SERVICE_NAME
    
    # 启动服务
    print_status "启动服务..."
    systemctl start $SERVICE_NAME
    
    # 检查服务状态
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "服务安装并启动成功！"
        print_success "服务地址: http://127.0.0.1:8080"
        print_success "Web界面: http://服务器IP:8080"
    else
        print_error "服务启动失败"
        systemctl status $SERVICE_NAME
        exit 1
    fi
}

uninstall_service() {
    print_status "开始卸载 Git代码文件代理服务..."
    
    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_status "停止服务..."
        systemctl stop $SERVICE_NAME
    fi
    
    # 禁用服务
    if systemctl is-enabled --quiet $SERVICE_NAME; then
        print_status "禁用服务..."
        systemctl disable $SERVICE_NAME
    fi
    
    # 删除服务文件
    if [ -f "$SERVICE_DIR/$SERVICE_NAME.service" ]; then
        print_status "删除服务文件..."
        rm -f $SERVICE_DIR/$SERVICE_NAME.service
    fi
    
    # 删除可执行文件
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_status "删除可执行文件..."
        rm -f $INSTALL_DIR/$BINARY_NAME
    fi
    
    # 重新加载systemd
    systemctl daemon-reload
    
    print_success "服务卸载完成！"
    print_warning "工作目录 $WORK_DIR 和日志目录 $LOG_DIR 保留，如需删除请手动删除"
}

show_status() {
    print_status "Git代码文件代理服务状态:"
    echo ""
    systemctl status $SERVICE_NAME --no-pager
    echo ""
    print_status "服务信息:"
    echo "  服务名称: $SERVICE_NAME"
    echo "  可执行文件: $INSTALL_DIR/$BINARY_NAME"
    echo "  工作目录: $WORK_DIR"
    echo "  日志目录: $LOG_DIR"
    echo "  服务地址: http://127.0.0.1:8080"
}

show_logs() {
    print_status "显示服务日志 (按Ctrl+C退出):"
    journalctl -u $SERVICE_NAME -f
}

show_help() {
    echo "Git代码文件代理服务安装脚本"
    echo ""
    echo "用法: $0 {install|uninstall|status|logs|start|stop|restart}"
    echo ""
    echo "命令说明:"
    echo "  install     - 安装服务"
    echo "  uninstall   - 卸载服务" 
    echo "  status      - 查看服务状态"
    echo "  logs        - 查看服务日志"
    echo "  start       - 启动服务"
    echo "  stop        - 停止服务"
    echo "  restart     - 重启服务"
    echo ""
    echo "服务信息:"
    echo "  端口: 8080"
    echo "  支持平台: GitHub, GitLab, Hugging Face, SourceForge"
    echo "  Web界面: http://服务器IP:8080"
}

# 检查root权限
check_root

# 处理参数
case "$1" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    start)
        print_status "启动服务..."
        systemctl start $SERVICE_NAME
        print_success "服务已启动"
        ;;
    stop)
        print_status "停止服务..."
        systemctl stop $SERVICE_NAME
        print_success "服务已停止"
        ;;
    restart)
        print_status "重启服务..."
        systemctl restart $SERVICE_NAME
        print_success "服务已重启"
        ;;
    *)
        show_help
        exit 1
        ;;
esac
