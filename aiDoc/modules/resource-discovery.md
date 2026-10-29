# 多来源资源检索

## 目标

让用户描述资源主题、用途和约束，由后续 Agent 自动选择 Telegram、GitHub 等来源完成检索、候选排序、下载和溯源，而不是要求用户预先知道精确文件名或指定检索渠道。

## 当前边界

- GVA Web 只访问 Go Server，不直连 `gva-telegram`。
- Xy 插件在 Go Server 内提供受 JWT 与 Casbin 保护的透明代理 `/xy/proxy/telegram/*`；星号后的方法、路径、查询、状态码和响应体保持 `gva-telegram` 原始契约。
- 闲鱼插件的所有 HTTP 入口统一使用 `/xy/*`。
- `gva-telegram` 继续负责 Telegram 登录、扫描、群组管理、资源检索和下载，原 Python 页面保留作为本地诊断入口。
- 当前 GVA 菜单把 Telegram 能力拆为概览与登录、群组与频道、资源检索、资源文件、服务设置五个页面。

## 后续 Agent 设计

不要先创建一个臆测式“万能搜索框架”。当 GitHub 来源开始实现时，围绕下面三个真实动作形成稳定契约：

1. `search`：输入自然语言意图、资源类型、数量和许可约束，返回轻量候选，不立即下载。
2. `inspect`：读取候选元数据、README、发布说明、许可证、大小和来源可信度，形成可解释评分。
3. `acquire`：用户或 Agent 选定候选后下载，计算哈希并保存来源、版本、提交 SHA、许可证和获取时间。

Agent 只编排动作和排序，不保存 Telegram Session 或 GitHub Token。各来源服务各自拥有凭证并返回统一的候选字段：`source`、`title`、`url`、`summary`、`resource_type`、`license`、`version`、`size`、`published_at`、`provenance`。

## 前端规划

- 当前阶段：在“Telegram 资源”子菜单内完成 Telegram 运维与检索闭环。
- GitHub 来源落地时：新增独立的“资源中心”，至少包含统一检索、任务记录、资源库、来源设置；Telegram 和 GitHub 成为来源设置下的并列渠道。
- 搜索结果必须展示来源、许可证、版本/提交、文件大小和下载状态；来源筛选只是偏好，默认允许 Agent 自动选择多个来源。

## 后端规划

- Go Server 负责 GVA 用户鉴权、任务所有权、统一任务/候选/资源记录和 Agent 工具入口。
- Telegram 保持 Python 独立服务，由 Xy 插件通过 `/xy/proxy/telegram/*` 提供管理 UI 所需的原始接口；Agent 侧通过受控工具调用，不让浏览器持有 Telegram 凭证。
- GitHub 来源优先直接使用 GitHub REST API。公开仓库检索和公开归档下载可以匿名进行，但匿名额度低；代码检索、私有仓库和稳定自动化应配置服务端凭证。
- 生产凭证优先使用 GitHub App 安装令牌；个人环境可使用最小权限 fine-grained token。Token 只存后端密钥配置，不进入前端、日志或任务表。
- 下载公开源码归档时固定 commit SHA；不要只保存可移动的分支名或 tag。下载后记录 SHA-256，便于复现与审计。

## 安全与合规

- 检索结果不等于允许分发。下载前检查仓库许可证、发布资产许可和 Telegram 内容来源授权。
- 禁止把 GVA JWT、用户 ID、Cookie 或 Authorization 头转发给下游服务；Go 代理只注入内部服务令牌。
- 归档解压沿用路径穿越、危险扩展名、大小和数量限制；外部资源进入资源库前必须计算哈希并保留来源记录。
- GitHub 限流必须遵守 `Retry-After` 与 `X-RateLimit-*`，不做并发暴力重试。
