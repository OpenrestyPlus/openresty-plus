# 节点部署与发布脚本

当前 Compose 提供三个联调节点：`openresty-east-1`、`openresty-east-2`、`openresty-east-3`。宿主机健康端口分别为 18080、18180、18280，Control API 转发端口分别为 18081、18181、18281；后两个节点另将容器业务端口 18082 映射到 18182、18282。每个节点使用独立的配置与日志目录。Go 控制面将候选写入各节点共享目录，在对应容器内执行 `nginx -t`，原子切换活动目录并 reload，最后查询各自工作进程的版本探针。三个节点均支持本地发布、权重分批与历史回滚。旧 UUID 自动登记接口仍独立存在。Filebeat 采集 east-1 日志并发送至 Kafka；新前端日志页目前直接读取本地容器及访问日志。

在控制面和三个节点启动后，从后端项目根目录运行 `python3 deploy/scripts/register-local-openresty-nodes.py`，将它们登记为新控制台的数字节点。多个 Center 时加 `--center-id`；脚本会拒绝已有节点身份或连接信息不一致的情况。

Compose 同时提供本地 MySQL、Redis 与 Kafka。Go 控制面启动时会自动应用仓库中的版本化 SQL，
在空数据库中创建表结构。若还要把旧外部 MySQL 的数据导入本地 MySQL，需在首次启动控制面前，
设置指向外部源库的 `MIGRATION_SOURCE_DB_*` 后运行
`./deploy/scripts/migrate-external-mysql-to-compose.sh`。该脚本只读源库，导入完成后逐表
校验行数，并在目标库已有表时拒绝覆盖。脚本会检查源库非空、导出文件包含每张表的建表语句，
并逐个确认目标表已创建，避免空导出或漏表时误报迁移成功。源库不能指向本地目标库；
服务的 `OPENRESTY_DB_*` 配置指向本地库时，不能用作外部迁移源。

本地 Docker 节点已经支持候选预检、全量与权重分批发布、金丝雀观察和历史快照回滚。生产节点适配与跨进程安全续跑仍需实现；当前本地适配器不得用于生产。

生产节点上的脚本必须由 root 持有、不可由 SSH 发布用户修改，并通过 `ForceCommand` 或受限 sudo 仅暴露固定动作。

生产脚本仍是待接入的前置文件，当前 Go 本地适配器没有调用它。生产实现须校验制品摘要、在非活动目录运行目标节点 `nginx -t`、原子切换活动目录，并通过本机 Unix socket 调用 Nginx Control API 的 REST reload 接口。

默认拒绝任意 shell、任意路径、未校验参数和 Control API 不可用时的信号回退。
