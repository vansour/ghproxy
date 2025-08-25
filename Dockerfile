# 多阶段构建Dockerfile
# 第一阶段：构建阶段
FROM golang:trixie AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o ghproxy main.go

# 第二阶段：运行阶段
FROM debian:trixie-slim

# 安装必要的包
RUN apt-get update && apt-get install -y ca-certificates wget && rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/ghproxy .

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./ghproxy"]
