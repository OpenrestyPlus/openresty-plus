# ORP Backend Development Notes

## 项目概述

本目录是独立的 Go 控制面项目，同时保存后端部署资产、平台文档和本地 Compose 集成环境。前端和节点 Agent 分别维护在相邻的独立项目目录中。

## 开发环境

- Go：1.26.1
- Node：22.18+ 或 24+
- pnpm：11.16.0
- Docker Compose：本地 MySQL、Redis、Kafka 与联调节点

## 常用命令

### 控制面

```bash
cd orp-backend
go test ./...
go run ./cmd/control-plane
```

控制面默认监听 `:8081`，数据库连接通过 `OPENRESTY_DB_*` 环境变量配置。平台级 Compose 文件位于本项目根目录。

### 前端

```bash
cd ../orp-frontend
pnpm install
pnpm -F @vben/web-antd run dev
pnpm -F @vben/web-antd run build
pnpm -F @vben/web-antd run typecheck
```

开发服务器将 `/api` 代理到 `http://127.0.0.1:8081`。

### Docker Compose 联调

```bash
cp .env.example .env
docker compose up -d mysql redis kafka
```

多项目完整联调从本项目目录运行 `./dev.sh`。本项目 Compose 提供 MySQL、Redis、Kafka、三个 OpenResty 节点和 Filebeat。

## 后端结构

- `cmd/control-plane`：控制面入口。
- `internal/config`：环境变量与数据库连接配置。
- `internal/store`：MySQL 连接。
- `internal/httpapi`：REST 路由、资源处理器、认证兼容接口与审计。

## 约束

- MySQL 是配置权威；Redis 和 Kafka 为本地联调依赖，不得阻断基础配置管理。
- 写操作必须记录审计事件；UUID 按 MySQL `BINARY(16)` 存储并在 API 层转换。
- 页面请求统一使用 `/api`；生产前端通过 Nginx 同源代理访问控制面。
- 保存配置不等同于节点生效；发布、原生配置物化和 reload 必须分别验证。
- 真实凭据仅通过环境变量注入，禁止写入源码、文档或前端构建产物。
