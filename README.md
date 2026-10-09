# OpenResty Plus

OpenResty Plus 是 OpenResty 集群管理平台。后端和前端源码分别保存在 `orp-backend/` 与 `orp-frontend/`，发布时先编译 Vue 管理界面，再将静态资源嵌入 Go 程序，最终交付一个同时提供管理界面和 API 的可执行文件。

## 构建单文件程序

构建环境要求：Go 1.26.1、Node.js 22.18+ 或 24.12+、pnpm 11.16.0。首次构建会自动安装前端依赖；完成后运行：

```sh
./build.sh
```

生成的程序位于 `dist/openresty-plus`。发布时只需复制这个可执行文件和配置文件；前端页面已嵌入程序，无需单独部署静态文件。
默认生成当前操作系统和 CPU 架构的程序。为 Linux x86-64 服务器构建时使用 `GOOS=linux GOARCH=amd64 ./build.sh`；ARM64 Linux 可使用 `GOOS=linux GOARCH=arm64 ./build.sh`。

## 配置及启动

在仓库根目录创建本地配置并填写数据库、管理员密码和数据加密密钥：

```sh
cp .env.example .env
```

程序启动时会自动读取当前目录的 `.env`。MySQL 是必需的外部服务；Redis 和 Kafka 不可用时，控制面仍可启动，但相关缓存或日志能力会降级。可用仓库中的 Compose 启动本地依赖：

```sh
docker compose --env-file .env -f orp-backend/docker-compose.yaml --project-directory orp-backend up -d mysql redis kafka
```

使用同一个可执行文件管理服务：

```sh
./dist/openresty-plus start       # 后台启动
./dist/openresty-plus status      # 查看状态
./dist/openresty-plus logs -f     # 实时查看日志
./dist/openresty-plus restart     # 重启
./dist/openresty-plus stop        # 优雅停止
./dist/openresty-plus version     # 查看版本
```

默认监听 `:8081`，打开 <http://127.0.0.1:8081> 访问管理界面。前台运行可使用 `./dist/openresty-plus run`，适合由 systemd、容器等进程管理器托管。PID 文件和日志默认写入仓库根目录的 `runtime/`；设置 `OPENRESTY_STATE_DIR` 可指定其他目录。

## 源码开发

单独启动前端开发服务器：

```sh
cd orp-frontend
pnpm install --frozen-lockfile
pnpm --filter @vben/web-antd dev
```

后端单独启动、环境变量和数据库迁移说明见 [`orp-backend/PROJECT.md`](orp-backend/PROJECT.md)。本地多节点 Compose 部署和其他平台文档见 [`orp-backend/deploy/README.md`](orp-backend/deploy/README.md) 与 [`orp-backend/docs/`](orp-backend/docs/)；前端文档见 [`orp-frontend/PROJECT.md`](orp-frontend/PROJECT.md)。

## 项目结构与许可

```text
.
├── orp-backend/    # Go 控制面及部署文件
└── orp-frontend/   # Vue 管理控制台和 pnpm 工作区
```

仓库许可证见 [`LICENSE`](LICENSE)。
