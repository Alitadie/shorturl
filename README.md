# ShortURL - High Performance URL Shortener Service

[![Go Backend CI](https://github.com/YOUR_GITHUB_USERNAME/shorturl/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_GITHUB_USERNAME/shorturl/actions)
![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![License](https://img.shields.io/badge/License-MIT-green)

A scalable URL shortener service built with Golang, Redis, SQLite (with Bloom Filter support). Designed using Domain-Driven Design (DDD) principles and Cache-Aside pattern.

## 🚀 Features

- **High Performance**: In-memory **Bloom Filter** to block malicious non-existent keys (Cache Penetration Protection).
- **Scalable ID**: **Base62** algorithm ensuring unique and non-colliding short links.
- **Cache Strategy**: Redis **Cache-Aside** pattern + Hotspot invalidation strategy.
- **Architecture**: 12-Factor App compliant, Clean Architecture (Handler -> Service -> Repository).
- **Deployment**: Dockerized & Cloud-Native ready (Docker Compose support).

## 🛠️ Architecture

`User -> [Nginx] -> Go App -> [Bloom Filter] -> Redis -> SQLite`

## 📦 Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose

### Quick Run (Docker)

```bash
# Clone the repo
git clone https://github.com/YOUR_GITHUB_USERNAME/shorturl.git
cd shorturl

# Start services
make docker-up
```

Access the service at: `http://localhost:8080`

### API Usage

**1. Create Short Link**

```bash
curl -X POST http://localhost:8080/shorten \
-H "Content-Type: application/json" \
-d '{"url": "https://www.google.com"}'
```

**2. Redirect**

```bash
curl -I http://localhost:8080/{short_id}
```

## 🧪 Testing

```bash
go test ./...
```

## 📄 License

MIT

