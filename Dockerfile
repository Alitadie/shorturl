# 阶段 1: 构建 (Builder)
# 使用带有 Go 编译器的官方镜像
FROM golang:1.25.4-alpine AS builder

# 设置工作目录
WORKDIR /app

# 先拷贝依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
# 下载依赖 (国内环境可能会慢，配置代理见下文)
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 拷贝源代码
COPY . .

# 编译 CGO_ENABLED=1 因为 SQLite 需要 C 编译器 (Alpine自带musl-gcc)
# 如果不想处理 gcc 问题，许多云原生方案会换成 MySQL/PG，这里我们继续用 SQLite 必须开启 CGO
RUN apk add --no-cache build-base
RUN CGO_ENABLED=1 GOOS=linux go build -o shorturl-server main.go

# --------------------------

# 阶段 2: 运行 (Runner)
# 使用极小的 Alpine 镜像 (只有 ~5MB)
FROM alpine:latest

WORKDIR /root/

# 安装必要的库 (SQLite依赖库) 和时区数据
RUN apk add --no-cache sqlite-libs tzdata

# 从第一阶段拷贝编译好的二进制文件
COPY --from=builder /app/shorturl-server .

# 创建一个数据目录用来挂载 SQLite 文件
RUN mkdir -p /data

# 设置环境变量默认值
ENV DB_PATH=/data/shorturl.db
ENV REDIS_ADDR=redis:6379
ENV GIN_MODE=release

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./shorturl-server"]
