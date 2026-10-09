# ADR-0010：Control API 作为生产发布硬依赖

## 状态

已接受

## 决策

节点必须通过本机 Unix socket 使用 Nginx Control REST API 执行 reload；控制面通过节点登记的 HTTP 转发地址编排请求。Control API 不可用、权限错误、返回失败或结果超时时，发布失败或进入未知状态，不自动回退到 `nginx -s reload`。

## 选择理由

- 保持唯一、可审计的 reload 控制路径。
- 避免信号控制和 REST 控制的状态语义不一致。
- 让节点能力门禁在发布前发现问题。

## 结果

节点登记和发布前校验必须确认 Control API 版本、编译能力、socket 路径、权限和 reload endpoint。旧节点若需兼容，必须单独设计并显式标记另一种节点能力，不得隐式降级。
