# ADR-0007：MVP 使用 Redis Streams 执行异步任务

## 状态

已废止（当前实现由 Go 控制面直接编排）

## 决策

历史方案曾计划使用 Redis Streams Consumer Group。当前发布、配置渲染和 reload 编排由 Go 控制面直接执行，任务结果和审计持久化到 MySQL；Redis 仅作为可选依赖。

## 选择理由

- 复用已确定的 Redis 基础设施。
- 支持消费确认、Pending 消息恢复和有限重试。
- MVP 不额外引入 Kafka 或 RabbitMQ 运维成本。

## 约束

任务必须具备幂等 ID、中心锁和状态机；Redis 丢失不能导致 MySQL 审计和最终状态不可恢复。
