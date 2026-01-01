# 第一阶段：编译
FROM golang:1.23.5-alpine3.21 AS builder

WORKDIR /app
# 设置代理（如果需要）
# ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 静态编译，去除符号表(-s -w)减小体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pilipili_frontend main.go

# 第二阶段：运行
FROM alpine:3.21

WORKDIR /app
# 安装时区数据和证书
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/pilipili_frontend .

# 设置时区
ENV TZ=Asia/Shanghai

ENTRYPOINT ["./pilipili_frontend"]
CMD ["config.yaml"]
