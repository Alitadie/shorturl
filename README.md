# 🚀 ShortURL - 企业级微服务

![Build Status](https://github.com/Alitadie/shorturl/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.25+-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Docker](https://img.shields.io/badge/Docker-Supported-blue)

一个用 Go 编写的高性能、可扩展的短链接服务。
拥有 **Base62 算法**、防止缓存穿透的 **布隆过滤器 (Bloom Filter)**、**分布式追踪** 和 **优雅停机** 等特性。

---

## ✨ 核心特性

- **高性能**: 采用 **Redis Cache-Aside** 策略优化。
- **安全性**: 集成 **布隆过滤器 (Bloom Filter)**，拦截 99% 的恶意不存在 ID 请求。
- **可扩展性**: 数据库 ID 采用 **Base62** 编码（无冲突）。
- **多数据库支持**: 一行代码切换 **MySQL / PostgreSQL / SQLite**。
- **可观测性**: 结构化日志 (Zap) 配合 **TraceID**。
- **可靠性**: 符合 12-Factor App 原则，支持优雅停机，容器化部署。

---

## 🛠️ 配置 (环境变量)

| 变量名           | 默认值              | 说明                            |
|------------------|---------------------|---------------------------------|
| `DB_DRIVER`      | `sqlite`            | `sqlite`, `mysql`, `postgres`   |
| `DB_HOST`        | `localhost`         | 数据库主机                      |
| `DB_USER`        | `root`              | 数据库用户                      |
| `DB_PASSWORD`    | -                   | 数据库密码                      |
| `REDIS_ADDR`     | `localhost:6379`    | Redis 地址                      |

---

## 📁 项目结构

```
shorturl/
├── .github/workflows/     # CI 自动化配置
│   └── ci.yml
├── config/                # 配置管理 (DB工厂, Redis, ENV)
│   └── db.go
├── docs/                  # Swagger 自动生成的文档
│   ├── docs.go
│   └── swagger.json
├── handler/               # HTTP 接口层 (Gin Handler)
│   └── http_hdl.go
├── middleware/            # 中间件 (Logger, Recovery)
│   └── logger.go
├── model/                 # 数据库模型 (GORM Struct)
│   └── link.go
├── pkg/                   # 公共工具包
│   └── base62/            # 核心算法
│       ├── base62.go
│       └── base62_test.go
├── repository/            # 仓储层 (DB+Redis+BloomFilter)
│   └── link_repo.go
├── data/                  # 挂载目录 (放.db文件)
├── docker-compose.yml     # 容器编排
├── Dockerfile             # 镜像构建
├── go.mod
├── go.sum
├── main.go                # 入口文件
├── Makefile               # 构建命令
├── README.md              # 说明书
└── LICENSE                # 开源协议 (新增)
```
---

## 🚀 快速开始

### 使用 Docker (推荐)

```bash
# 1. 使用 Docker Compose 运行 (包含 Redis 和 应用)
make docker-up

# 2. (可选) 切换到 MySQL
# 在 docker-compose.yml 中取消注释 MySQL 部分并重启
```

### 本地开发

```bash
# 1. 启动依赖
docker run -d -p 6379:6379 redis:alpine

# 2. 运行应用
go run main.go
```

API 文档地址: `http://localhost:8080/swagger/index.html`

---

## 🔗 API 参考

**POST /shorten** - 创建短链接
```json
{"url": "https://www.google.com"}
```

**GET /:id** - 重定向
```bash
curl -I http://localhost:8080/AbC9
```

---

## 🤝 贡献

欢迎提交 Pull Request。对于重大更改，请先提交 Issue 进行讨论。
