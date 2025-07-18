# 构建阶段
FROM golang:alpine AS builder

WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache build-base git

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# 构建应用 (注意修正二进制文件名)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-service ./main.go

# 最终阶段
FROM alpine:3.21.3

# 安装SSL证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件 (修正名称)
COPY --from=builder /app/user-service .

# 暴露端口
EXPOSE 8080

# 设置默认环境变量
ENV DB_HOST=host.docker.internal \
    DB_PORT=3306 \
    DB_USER=root \
    DB_NAME=ecommerce \
    RABBITMQ_URL="amqp://admin:rabbitmq@IP:5672/" \
    SMTP_HOST=smtp.gmail.com \
    SMTP_PORT=587

# 使用非root用户运行
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./user-service"]
