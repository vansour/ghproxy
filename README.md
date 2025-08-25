# Git代码文件加速代理服务

[![Docker Hub](https://img.shields.io/docker/pulls/vansour/ghproxy.svg)](https://hub.docker.com/r/vansour/ghproxy)
[![Docker Image Size](https://img.shields.io/docker/image-size/vansour/ghproxy/latest)](https://hub.docker.com/r/vansour/ghproxy)
[![Docker Image Version](https://img.shields.io/docker/v/vansour/ghproxy/latest)](https://hub.docker.com/r/vansour/ghproxy)

一个高性能的Git代码文件加速代理服务，支持GitHub、GitLab、Hugging Face和SourceForge等多个平台的文件加速下载。

## 🚀 功能特性

- **多平台支持**: GitHub、GitLab、Hugging Face、SourceForge
- **智能转换**: 自动将blob链接转换为raw下载链接  
- **Git克隆加速**: 支持通过代理进行git clone操作
- **现代化界面**: 响应式Web界面，支持链接生成和一键复制
- **无超时限制**: 支持大文件和大型仓库的长时间传输
- **RESTful API**: 提供API接口用于自动化集成
- **多种部署**: 支持Docker、systemd等多种部署方式

## 📦 支持的平台

| 平台 | 域名 | 支持功能 |
|------|------|----------|
| **GitHub** | github.com | ✅ 文件下载 ✅ Git克隆 |
| **GitLab** | gitlab.com | ✅ 文件下载 ✅ Git克隆 |
| **Hugging Face** | huggingface.co | ✅ 文件下载 |
| **SourceForge** | sourceforge.net | ✅ 文件下载 |

## 安装使用

### 系统要求

- Linux系统
- Go 1.16+
- systemd支持
- root权限

## 🐳 快速开始

### 方式一：使用Docker Hub镜像（推荐）

```bash
# 拉取镜像
docker pull vansour/ghproxy:latest

# 运行容器
docker run -d --name ghproxy -p 8080:8080 vansour/ghproxy:latest

# 访问服务
# Web界面: http://localhost:8080
```

### 方式二：使用Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'
services:
  ghproxy:
    image: vansour/ghproxy:latest
    container_name: ghproxy
    ports:
      - "8080:8080"
    environment:
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

运行：
```bash
docker-compose up -d
```

### 方式三：本地构建Docker

#### 系统要求
- Docker 20.10+
- Docker Compose 2.0+

#### 构建步骤
```bash
# 1. 下载代码
git clone <repository-url>
cd ghproxy

# 2. 构建并启动服务
./docker.sh build
./docker.sh start

# 3. 访问服务
# Web界面: http://服务器IP:8080
```

## 📖 使用方法

### Web界面使用

访问 `http://localhost:8080` 打开Web界面，输入GitHub、GitLab等链接即可生成加速下载链接。

### 直接代理使用

将原始链接中的域名前加上代理地址：

```bash
# 原始链接
https://github.com/user/repo/blob/main/file.txt

# 代理链接
http://localhost:8080/https://github.com/user/repo/blob/main/file.txt
```

### Git克隆加速

```bash
git clone http://localhost:8080/https://github.com/user/repo.git
```

### API接口

生成加速链接：
```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://github.com/user/repo/blob/main/file.txt"}'
```

响应示例：
```json
{
  "success": true,
  "browser_link": "http://localhost:8080/https://github.com/user/repo/blob/main/file.txt",
  "wget_command": "wget \"http://localhost:8080/https://github.com/user/repo/blob/main/file.txt\" -O file.txt",
  "curl_command": "curl -L \"http://localhost:8080/https://github.com/user/repo/blob/main/file.txt\" -o file.txt",
  "git_command": "git clone http://localhost:8080/https://github.com/user/repo.git"
}
```

## 🔧 配置选项

### 环境变量

- `TZ`: 时区设置（默认: Asia/Shanghai）
- `PORT`: 服务端口（默认: 8080）

### Docker镜像版本

- `latest`: 最新稳定版本（无超时限制）
- `v1.1.0`: 无超时限制版本（推荐用于大文件传输）
- `v1.0.0`: 基础版本（有超时限制）

#### Docker管理命令
```bash
# 查看服务状态
./docker.sh status

# 查看日志
./docker.sh logs

# 重启服务
./docker.sh restart

# 停止服务
./docker.sh stop

# 更新服务
./docker.sh update

# 清理资源
./docker.sh cleanup
```

### 方式二：系统服务部署

#### 系统要求
- Linux系统
- Go 1.16+
- systemd支持
- root权限

### 安装步骤

1. **下载代码**
```bash
git clone <repository-url>
cd ghproxy
```

2. **安装服务**
```bash
sudo ./install.sh install
```

3. **访问服务**
- Web界面: http://服务器IP:8080
- 直接代理: http://服务器IP:8080/完整的文件URL

### 服务管理

```bash
# 查看服务状态
sudo ./install.sh status

# 启动服务
sudo ./install.sh start

# 停止服务
sudo ./install.sh stop

# 重启服务
sudo ./install.sh restart

# 查看实时日志
sudo ./install.sh logs

# 卸载服务
sudo ./install.sh uninstall
```

## 使用方法

### Web界面使用

1. 打开浏览器访问 `http://服务器IP:8080`
2. 在输入框中粘贴原始链接
3. 点击"生成加速链接"
4. 复制所需格式的链接使用

### 直接代理使用

将原始URL前面加上代理地址即可：

```bash
# 原始链接
https://github.com/user/repo/blob/main/file.txt

# 代理链接  
http://你的服务器:8080/https://github.com/user/repo/blob/main/file.txt
```

### 命令行使用

```bash
# wget下载
wget "http://你的服务器:8080/原始URL" -O 文件名

# curl下载
curl -L "http://你的服务器:8080/原始URL" -o 文件名

# git clone (仅支持仓库链接)
git clone http://你的服务器:8080/仓库URL.git
```

## 示例

```bash
# GitHub文件下载
wget "http://127.0.0.1:8080/https://github.com/golang/go/blob/master/README.md" -O README.md

# GitLab文件下载
curl -L "http://127.0.0.1:8080/https://gitlab.com/gitlab-org/gitlab/-/blob/master/README.md" -o README.md

# Hugging Face模型文件
wget "http://127.0.0.1:8080/https://huggingface.co/microsoft/DialoGPT-medium/resolve/main/config.json" -O config.json

# Git仓库克隆
git clone http://127.0.0.1:8080/https://github.com/golang/go.git
```

## 技术特性

- **日志管理**: 自动轮转，单文件限制5MB
- **服务管理**: systemd保活，开机自启
- **安全配置**: 非特权用户运行，安全沙箱
- **高可用**: 服务异常自动重启

## 开源协议

本项目采用 MIT 开源协议
