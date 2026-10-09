# ADR-0009：使用 rsync 与受限 SSH 固定脚本发布

## 状态

已接受

## 决策

生产部署可通过 rsync over SSH 传输制品；当前 Compose 联调使用项目目录绑定挂载。两种方式最终都由节点本机 Unix Socket Control API 执行 reload。

## 结果

控制面不执行任意远程 shell。节点 Control API 的 Unix Socket 不直接暴露给浏览器，联调环境通过节点登记的 HTTP 转发地址访问，生产环境应保留受限 SSH 和固定脚本边界。
