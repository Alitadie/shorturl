.PHONY: run build docker-up docker-down docker-build

# 本地运行 (需要本地跑着 Redis)
run:
	go run main.go

# 本地编译
build:
	go build -o bin/server main.go

# Docker 全套启动
docker-up:
	docker-compose up -d --build

# Docker 全套关闭
docker-down:
	docker-compose down

# 查看 Docker 日志
docker-logs:
	docker-compose logs -f
