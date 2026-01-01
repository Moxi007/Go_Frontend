# 第一阶段：编译
FROM golang:1.23.5-alpine3.21 AS builder

WORKDIR /app

# 【优化1】取消注释，启用国内代理，确保依赖能下载成功
# ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 【优化2】新增这一行：自动整理依赖
# 它会自动下载代码中引用但 go.mod 缺失的库 (比如 singleflight)
RUN go mod tidy

# 静态编译，去除符号表(-s -w)减小体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go_frontend main.go

# 第二阶段：运行
FROM alpine:3.21

WORKDIR /app

# 安装时区数据和证书 (HTTPS 请求必须要有 ca-certificates)
RUN apk --no-cache add ca-certificates tzdata

# 从第一阶段复制编译好的二进制文件
COPY --from=builder /app/go_frontend .

# 设置时区
ENV TZ=Asia/Shanghai

# 入口
ENTRYPOINT ["./go_frontend"]
CMD ["config.yaml"]
