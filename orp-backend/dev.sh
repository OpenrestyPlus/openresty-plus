#!/usr/bin/env bash
set -Eeuo pipefail

backend_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
workspace_root=$(cd "$backend_dir/.." && pwd)
cd "$workspace_root"
compose=(docker compose --env-file "$workspace_root/.env" -f "$backend_dir/docker-compose.yaml" --project-directory "$backend_dir")

usage() {
  printf '用法：%s [--fresh-build]\n' "$(basename "$0")"
  printf '默认使用 Go 增量编译；--fresh-build 会强制重新编译所有 Go 包。\n'
}

fresh_build=false
case "${1:-}" in
  "") ;;
  --fresh-build) fresh_build=true ;;
  -h|--help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

for command_name in go node pnpm docker curl; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '缺少命令：%s\n' "$command_name" >&2
    exit 1
  fi
done

node_is_supported() {
  node -e 'const [major, minor] = process.versions.node.split(".").map(Number); process.exit((major === 22 && minor >= 18) || major >= 24 ? 0 : 1)'
}

node_version=$(node -p 'process.versions.node')
if ! node_is_supported; then
  nvm_dir=${NVM_DIR:-"${HOME}/.nvm"}
  switched_node=false

  if [[ -s "$nvm_dir/nvm.sh" ]]; then
    # 非交互 bash 不会自动加载 nvm；仅在当前版本不兼容时加载并切换。
    source "$nvm_dir/nvm.sh"
    candidate_nodes=("$nvm_dir"/versions/node/v22.* "$nvm_dir"/versions/node/v24.*)
    for (( candidate_index=${#candidate_nodes[@]}-1; candidate_index>=0; candidate_index-- )); do
      version_dir=${candidate_nodes[$candidate_index]}
      [[ -x "$version_dir/bin/node" ]] || continue
      candidate_version=${version_dir##*/v}
      candidate_major=${candidate_version%%.*}
      candidate_minor=${candidate_version#*.}
      candidate_minor=${candidate_minor%%.*}
      if [[ "$candidate_major" == 24 || ( "$candidate_major" == 22 && "$candidate_minor" -ge 18 ) ]]; then
        if nvm use "$candidate_version" --silent >/dev/null 2>&1 && node_is_supported; then
          node_version=$(node -p 'process.versions.node')
          printf '已通过 nvm 自动切换到 Node.js %s。\n' "$node_version"
          switched_node=true
          break
        fi
      fi
    done
  fi

  if [[ "$switched_node" != true ]]; then
    printf '当前 Node.js 版本为 %s；前端要求 Node.js 22.18+ 或 24+。\n' "$node_version" >&2
    printf '请安装兼容版本后重试，例如：nvm install 24\n' >&2
    exit 1
  fi
fi

pnpm_version=$(pnpm --version) || {
  printf '无法通过 Corepack 启动项目指定的 pnpm；请使用兼容版本的 Node.js。\n' >&2
  exit 1
}
if [[ "$pnpm_version" != "11.16.0" ]]; then
  printf 'pnpm 版本为 %s，项目要求 11.16.0。\n' "$pnpm_version" >&2
  exit 1
fi

check_port_available() {
  local port=$1 service=$2
  local listener_pids=""
  if command -v lsof >/dev/null 2>&1; then
    listener_pids=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
  elif curl -fsS --max-time 1 "http://127.0.0.1:$port/" >/dev/null 2>&1; then
    listener_pids="unknown"
  fi

  if [[ -n "$listener_pids" ]]; then
    printf '端口 %s 已被占用（%s，PID: %s）。请先停止占用该端口的旧服务，再重新运行 ./dev.sh。\n' \
      "$port" "$service" "${listener_pids//$'\n'/, }" >&2
    exit 1
  fi
}

# 在编译和启动后端前检查，避免端口冲突时留下半启动状态。
check_port_available 8081 'Go 控制面'
check_port_available 5666 'Vite 前端'

if [[ ! -f .env ]]; then
  printf '未找到工作区根目录 .env。请先执行：cp .env.example .env，并填写本地 MySQL 配置。\n' >&2
  exit 1
fi

# 逐行加载 .env，避免 JDBC URL 中的 & 被 shell 当作控制符。
while IFS= read -r line || [[ -n "$line" ]]; do
  line=${line%$'\r'}
  [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
  [[ "$line" == *=* ]] || continue

  key=${line%%=*}
  value=${line#*=}
  [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || continue
  if [[ ${#value} -ge 2 && ( "${value:0:1}" == '"' || "${value:0:1}" == "'" ) ]]; then
    value=${value:1:${#value}-2}
  fi
  if ! declare -p "$key" >/dev/null 2>&1; then
    export "$key=$value"
  fi
done < "$workspace_root/.env"

export OPENRESTY_DB_URL=${OPENRESTY_DB_URL:-"jdbc:mysql://127.0.0.1:${MYSQL_PORT:-3306}/${MYSQL_DATABASE:-openresty-plus}?useUnicode=true&characterEncoding=utf8&serverTimezone=Asia/Shanghai"}
export OPENRESTY_DB_USERNAME=${OPENRESTY_DB_USERNAME:-${MYSQL_USER:-openresty-plus}}
export OPENRESTY_DB_PASSWORD=${OPENRESTY_DB_PASSWORD:-${MYSQL_PASSWORD:-}}
export OPENRESTY_HTTP_ADDR=${OPENRESTY_HTTP_ADDR:-:8081}

runtime_dir="$workspace_root/runtime/dev"
mkdir -p "$runtime_dir"
export OPENRESTY_ADMIN_USERNAME=${OPENRESTY_ADMIN_USERNAME:-vben}
if [[ -z "${OPENRESTY_ADMIN_PASSWORD:-}" ]]; then
  credentials_file="$runtime_dir/admin-password"
  if [[ ! -f "$credentials_file" ]]; then
    umask 077
    openssl rand -hex 18 > "$credentials_file"
  fi
  OPENRESTY_ADMIN_PASSWORD=$(<"$credentials_file")
  export OPENRESTY_ADMIN_PASSWORD
  printf '本地管理员账号：%s；密码保存在 runtime/dev/admin-password。\n' "$OPENRESTY_ADMIN_USERNAME"
fi
local_frontend_env="$workspace_root/orp-frontend/apps/web-antd/.env.local"
umask 077
printf 'VITE_LOCAL_DEV_PASSWORD=%s\n' "$OPENRESTY_ADMIN_PASSWORD" > "$local_frontend_env"
if [[ -z "${OPENRESTY_DATA_KEY:-}" ]]; then
  data_key_file="$runtime_dir/data-key"
  if [[ ! -f "$data_key_file" ]]; then
    umask 077
    openssl rand -hex 32 > "$data_key_file"
  fi
  OPENRESTY_DATA_KEY=$(<"$data_key_file")
  export OPENRESTY_DATA_KEY
fi

if [[ ! -x "$workspace_root/orp-frontend/node_modules/.bin/vite" ]]; then
  printf '前端依赖未安装，执行 pnpm install……\n'
  (cd "$workspace_root/orp-frontend" && pnpm install --frozen-lockfile)
fi

printf '编译 Go 控制面……\n'
if [[ "$fresh_build" == true ]]; then
  (cd "$backend_dir" && go build -a -trimpath -o "$runtime_dir/control-plane" ./cmd/control-plane)
else
  (cd "$backend_dir" && go build -trimpath -o "$runtime_dir/control-plane" ./cmd/control-plane)
fi

printf '启动本地 MySQL、Redis、Kafka……\n'
# Remove containers left behind by older Compose files so their ports do not hide local rebuilds.
for old_container in openresty-plus-control-plane openresty-plus-frontend; do
  if docker container inspect "$old_container" >/dev/null 2>&1; then
    docker stop "$old_container" >/dev/null
  fi
done
"${compose[@]}" up -d mysql redis kafka

printf '等待 MySQL 就绪……\n'
mysql_ready=false
for _ in {1..60}; do
  if "${compose[@]}" exec -T mysql mysqladmin ping -h 127.0.0.1 \
    -uroot -p"${MYSQL_ROOT_PASSWORD:-}" --silent >/dev/null 2>&1; then
    mysql_ready=true
    break
  fi
  sleep 2
done
if [[ "$mysql_ready" != true ]]; then
  printf 'MySQL 未能在 120 秒内就绪，请检查：docker compose -f orp-backend/docker-compose.yaml logs mysql\n' >&2
  exit 1
fi

if curl -fsS http://127.0.0.1:8081/healthz >/dev/null 2>&1; then
  printf '端口 8081 已有控制面在响应，请先关闭它后重试，避免误连旧进程。\n' >&2
  exit 1
fi

printf '启动 Go 控制面（%s），日志：runtime/dev/control-plane.log\n' "$OPENRESTY_HTTP_ADDR"
"$runtime_dir/control-plane" >>"$runtime_dir/control-plane.log" 2>&1 &
backend_pid=$!

cleanup() {
  if kill -0 "$backend_pid" 2>/dev/null; then
    kill "$backend_pid" 2>/dev/null || true
    wait "$backend_pid" 2>/dev/null || true
  fi
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

printf '等待控制面完成数据库表结构初始化/升级……\n'
backend_ready=false
for _ in {1..60}; do
  if ! kill -0 "$backend_pid" 2>/dev/null; then
    printf 'Go 控制面启动失败，最近日志如下：\n' >&2
    tail -n 80 "$runtime_dir/control-plane.log" >&2
    exit 1
  fi
  if curl -fsS http://127.0.0.1:8081/healthz >/dev/null 2>&1; then
    backend_ready=true
    break
  fi
  sleep 1
done
if [[ "$backend_ready" != true ]]; then
  printf 'Go 控制面未能在 60 秒内就绪，日志：runtime/dev/control-plane.log\n' >&2
  exit 1
fi

printf '启动 Filebeat 日志采集并连接 Kafka……\n'
"${compose[@]}" up -d filebeat-east-1 filebeat-east-2 filebeat-east-3

printf '启动 Vite 前端开发服务器：http://127.0.0.1:5666\n'
printf 'API：http://127.0.0.1:8081/api（前端通过 Vite /api 代理访问）\n'
printf '按 Ctrl+C 停止前端和 Go 控制面；数据库与中间件容器保持运行。\n'
cd "$workspace_root/orp-frontend"
pnpm --filter @vben/web-antd run dev
