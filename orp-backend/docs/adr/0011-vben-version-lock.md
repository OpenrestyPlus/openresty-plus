# ADR-0011：锁定 Vben Admin v5.7.0 与匹配工具链

## 状态

已接受

## 决策

前端使用官方 Vben Admin `v5.7.0` tag，Node.js 使用 22.18.0 LTS，pnpm 使用该 tag 声明的 10.33.4。

## 背景

官方仓库当前 `main` 与固定 tag 的 package manager 声明不同。为保证可复现构建，工具版本必须与固定 tag 一致。

## 结果

升级 Vben 时必须同时评估 tag、Node 和 pnpm 版本；禁止仅升级包管理器而保留旧模板代码。
