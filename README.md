# ShortURL - 高性能 URL 短链接服务

[![Go Backend CI](https://github.com/Alitadie/shorturl/actions/workflows/ci.yml/badge.svg)](https://github.com/Alitadie/shorturl/actions)
![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![License](https://img.shields.io/badge/License-MIT-green)

基于 Golang、Redis、SQLite（支持布隆过滤器）构建的可扩展 URL 短链接服务。采用领域驱动设计 (DDD) 原则和 Cache-Aside 模式设计。

## 🚀 功能特性

- **高性能**: 内存 **布隆过滤器** 拦截恶意不存在的 Key（防止缓存穿透）。
- **可扩展 ID**: **Base62** 算法确保生成唯一且不冲突的短链接。
- **缓存策略**: Redis **Cache-Aside** 模式 + 热点失效策略。
- **架构设计**: 符合 12-Factor App 标准，整洁架构 (Handler -> Service -> Repository)。
- **部署**: 容器化 & 云原生就绪 (支持 Docker Compose)。

## 🛠️ 架构设计

`User -> [Nginx] -> Go App -> [Bloom Filter] -> Redis -> SQLite`

## 📦 快速开始

### 环境要求

- Go 1.25.4+
- Docker & Docker Compose

### 快速运行 (Docker)

```bash
# 克隆仓库
git clone https://github.com/Alitadie/shorturl.git
cd shorturl

# 启动服务
make docker-up
```

服务访问地址: `http://localhost:8080`

### API 使用指南

**1. 创建短链接**

```bash
curl -X POST http://localhost:8080/shorten \
-H "Content-Type: application/json" \
-d '{"url": "https://www.google.com"}'
```

**2. 重定向**

```bash
curl -I http://localhost:8080/{short_id}
```

## 🧪 测试

```bash
go test ./...
```

## 📄 许可证

MIT

