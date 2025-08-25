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

### 方式三：一键安装到服务器（systemd）

```bash
# 一键安装命令
wget https://raw.githubusercontent.com/vansour/ghproxy/main/install.sh -O ghproxy.sh && chmod +x ./ghproxy.sh && ./ghproxy.sh
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