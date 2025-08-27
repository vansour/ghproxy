# 多阶段构建Dockerfile - 支持多架构
# 第一阶段：构建阶段
FROM --platform=$BUILDPLATFORM golang:trixie AS builder

# 构建参数
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG BUILDTIME=unknown

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go ./

# 构建应用（支持多架构）
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILDTIME}" \
    -o ghproxy main.go

# 第二阶段：运行阶段
FROM --platform=$TARGETPLATFORM debian:trixie-slim

# 标签信息
LABEL org.opencontainers.image.title="Git代码文件加速代理服务" \
      org.opencontainers.image.description="支持GitHub、GitLab、Hugging Face、SourceForge的多平台代理服务" \
      org.opencontainers.image.vendor="vansour" \
      org.opencontainers.image.source="https://github.com/vansour/ghproxy" \
      org.opencontainers.image.documentation="https://github.com/vansour/ghproxy/blob/main/README.md"

# 安装必要的包
RUN apt-get update && apt-get install -y ca-certificates wget curl && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/ghproxy .

# 确保可执行文件有执行权限
RUN chmod +x ./ghproxy

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 启动应用
CMD ["./ghproxy"]
