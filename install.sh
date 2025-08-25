#!/bin/bash

# Git代码文件加速代理服务一键安装脚本
# 支持 GitHub、GitLab、Hugging Face、SourceForge
# 
# 使用方法:
# wget https://raw.githubusercontent.com/vansour/ghproxy/main/install.sh -O ghproxy.sh && chmod +x ./ghproxy.sh && ./ghproxy.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SERVICE_NAME="ghproxy"
BINARY_NAME="ghproxy"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
WORK_DIR="/opt/ghproxy"
LOG_DIR="/var/log/ghproxy"
GITHUB_REPO="https://github.com/vansour/ghproxy"
TEMP_DIR="/tmp/ghproxy-install"

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
        print_status "请使用: sudo $0"
        exit 1
    fi
}

check_dependencies() {
    print_status "检查依赖..."
    
    # 检查必要的命令
    for cmd in wget git go; do
        if ! command -v $cmd &> /dev/null; then
            print_error "$cmd 未安装"
            case $cmd in
                wget)
                    print_status "请安装wget: apt-get install wget 或 yum install wget"
                    ;;
                git)
                    print_status "请安装git: apt-get install git 或 yum install git"
                    ;;
                go)
                    print_status "请安装Go语言环境: https://golang.org/dl/"
                    ;;
            esac
            exit 1
        fi
    done
    print_success "依赖检查通过"
}

download_source() {
    print_status "下载源代码..."
    
    # 清理临时目录
    rm -rf $TEMP_DIR
    mkdir -p $TEMP_DIR
    cd $TEMP_DIR
    
    # 克隆仓库
    git clone $GITHUB_REPO.git .
    if [ $? -ne 0 ]; then
        print_error "下载源代码失败"
        exit 1
    fi
    
    print_success "源代码下载完成"
}

install_service() {
    print_status "开始安装 Git代码文件加速代理服务..."
    
    # 进入源代码目录
    cd $TEMP_DIR
    
    # 编译程序
    print_status "编译程序..."
    go mod tidy
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
    cat > $SERVICE_DIR/$SERVICE_NAME.service << 'EOF'
[Unit]
Description=Git代码文件加速代理服务
Documentation=https://github.com/vansour/ghproxy
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
WorkingDirectory=/opt/ghproxy
ExecStart=/usr/local/bin/ghproxy
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/ghproxy

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
    
    # 重新加载systemd配置
    print_status "重新加载systemd配置..."
    systemctl daemon-reload
    
    # 启用并启动服务
    print_status "启用并启动服务..."
    systemctl enable $SERVICE_NAME
    systemctl start $SERVICE_NAME
    
    # 清理临时文件
    print_status "清理临时文件..."
    cd /
    rm -rf $TEMP_DIR
    
    # 检查服务状态
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "Git代码文件加速代理服务安装完成！"
        echo ""
        print_status "服务信息:"
        echo "  服务名称: $SERVICE_NAME"
        echo "  服务地址: http://127.0.0.1:8080"
        echo "  服务状态: $(systemctl is-active $SERVICE_NAME)"
        echo ""
        print_status "常用命令:"
        echo "  查看状态: systemctl status $SERVICE_NAME"
        echo "  查看日志: journalctl -u $SERVICE_NAME -f"
        echo "  重启服务: systemctl restart $SERVICE_NAME"
        echo "  停止服务: systemctl stop $SERVICE_NAME"
        echo ""
        print_success "现在可以通过 http://127.0.0.1:8080 访问服务！"
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
    echo "Git代码文件加速代理服务一键安装脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令说明:"
    echo "  install     - 安装服务 (默认)"
    echo "  uninstall   - 卸载服务" 
    echo "  status      - 查看服务状态"
    echo "  logs        - 查看服务日志"
    echo "  start       - 启动服务"
    echo "  stop        - 停止服务"
    echo "  restart     - 重启服务"
    echo "  help        - 显示帮助信息"
    echo ""
    echo "一键安装命令:"
    echo "  wget https://raw.githubusercontent.com/vansour/ghproxy/main/install.sh -O ghproxy.sh && chmod +x ./ghproxy.sh && ./ghproxy.sh"
    echo ""
    echo "服务信息:"
    echo "  端口: 8080"
    echo "  支持平台: GitHub, GitLab, Hugging Face, SourceForge"
    echo "  Web界面: http://服务器IP:8080"
}

# 主函数
main() {
    case "${1:-install}" in
        install)
            check_root
            check_dependencies
            download_source
            install_service
            ;;
        uninstall)
            check_root
            uninstall_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        start)
            check_root
            systemctl start $SERVICE_NAME
            print_success "服务已启动"
            ;;
        stop)
            check_root
            systemctl stop $SERVICE_NAME
            print_success "服务已停止"
            ;;
        restart)
            check_root
            systemctl restart $SERVICE_NAME
            print_success "服务已重启"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
