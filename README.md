# OpenResty Plus

OpenResty Plus 是由 Go 控制面和 Vue 管理控制台组成的 OpenResty 集群管理平台。本仓库将原先独立维护的后端与前端项目放在同一仓库中，两个项目仍保留各自的依赖、构建和部署配置。

## 项目结构

```text
.
├── orp-backend/    # Go 控制面、数据库迁移、Compose 部署及后端文档
└── orp-frontend/   # Vue 3 管理控制台、pnpm workspace 及前端文档
```

## 快速开始

### 启动后端

后端要求 Go 1.26.1。单独运行、环境变量和数据库初始化说明见 [`orp-backend/PROJECT.md`](orp-backend/PROJECT.md)。使用本地 Compose 启动控制面及依赖服务：

```sh
cd orp-backend
docker compose up -d
```

完整部署说明和本地三节点配置见 [`orp-backend/deploy/README.md`](orp-backend/deploy/README.md)。

### 启动前端

前端要求 Node.js 22.18+ 或 24.12+，以及仓库指定的 pnpm 11.16.0。在仓库根目录执行：

```sh
cd orp-frontend
pnpm install --frozen-lockfile
pnpm --filter @vben/web-antd dev
```

前端环境配置、构建及其他应用说明见 [`orp-frontend/PROJECT.md`](orp-frontend/PROJECT.md)。

## 联合开发

后端 `dev.sh` 可用于本地多服务联调。请先查看 [`orp-backend/PROJECT.md`](orp-backend/PROJECT.md) 和脚本说明，准备所需环境配置后，从 `orp-backend/` 目录运行：

```sh
cd orp-backend
./dev.sh
```

该脚本会从工作区根目录读取 `.env`，启动本地 MySQL、Redis、Kafka 及控制面，并启动 `orp-frontend/` 中的 Vite。请先在仓库根目录准备 `.env`（可参考 `orp-backend/.env.example`），并安装 Docker Compose、Go 1.26.1、Node.js 22.18+ 或 24.12+、pnpm 11.16.0。无需完整联调时，可按上面的步骤分别启动后端和前端。

## 文档与许可

- 后端能力与限制：[`orp-backend/README.md`](orp-backend/README.md)
- 前端项目说明：[`orp-frontend/README.md`](orp-frontend/README.md)
- 后端架构、部署与功能文档：[`orp-backend/docs/`](orp-backend/docs/)
- 前端工作区文档：[`orp-frontend/README.zh-CN.md`](orp-frontend/README.zh-CN.md)
- 仓库许可证：[`LICENSE`](LICENSE)
