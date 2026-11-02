# Telegram Analyzer & Cleaner

本地、单用户的 Telegram 群聊/频道分析、手工清理与授权资源检索工具。资源检索只读取已加入并已扫描的公开正常来源，不会自动加入群组。

## 安装与启动

1. 在 [my.telegram.org](https://my.telegram.org) 创建应用，获取 API ID 与 API Hash。
2. 配置并启动：

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# 编辑 .env
python3 app.py
```

如果当前网络需要 SOCKS5 代理，在 `.env` 中配置：

```env
TG_PROXY_HOST=127.0.0.1
TG_PROXY_PORT=17891
```

通过 Docker Compose 运行且代理位于宿主机时，将 `TG_PROXY_HOST` 设置为 `host.docker.internal`。

浏览器打开 <http://127.0.0.1:8080>。首次使用点击“生成二维码”，用手机 Telegram 的“设置 → 设备 → 链接桌面设备”扫码。登录 session 保存在 PostgreSQL `xianyu_assistant.tel_settings` 中，应按凭据保护，不要导出或分享。

Streamable HTTP MCP 端点为 `/mcp/`，固定提供渠道、来源、资源和通用任务工具。调用必须携带 `x-internal-token`，其值来自 `TELEGRAM_INTERNAL_TOKEN`；未配置令牌时 MCP 端点拒绝服务。

## 使用

- 快速扫描只遍历 dialogs，保存名称、类型、人数、最新消息时间、未读、静音、归档和创建者状态。
- 只有手动创建资源任务后才会在所选来源中搜索消息文本和文件名，匹配扩展名后下载；并发默认 1，最大 3。
- 群聊页一次展示全部结果，集成名称模糊搜索、筛选、升降序、Telegram 跳转、勾选退出和 CSV 导出。
- 退出前会再弹窗确认；每项结果写入 `leave_history`。

运行元数据自检；设置 `DATABASE_URL` 时还会在共用 PostgreSQL 中运行隔离的增删改查集成检查（仅操作唯一测试数据）：

```bash
python test_core.py
```

局域网访问会暴露管理页面；仅在可信网络使用，并用 macOS 防火墙限制访问。Bootstrap CSS 使用 CDN，离线时功能仍可用但样式会变简陋。

## Docker 与 PostgreSQL

Compose 中的服务统一使用 `go.xy.assistant` 网络。闲鱼助手后台映射到宿主机 `18080`，PostgreSQL 映射到 `15432`，RustFS 的 S3 API/控制台分别映射到 `19000`/`19001`，Telegram Web UI 映射到 `127.0.0.1:18081`。

Telegram 与 Go 后端共用 `xianyu_assistant_v2` 数据库，Telegram 表统一使用 `tel_` 前缀；PostgreSQL 通过宿主机 `15432` 端口连接。原 SQLite 业务数据不再使用，不提供 SQLite 迁移流程。

资源任务表为 `tel_resource_tasks`、`tel_resource_files`、`tel_resource_sources`。文件临时落在项目根目录 `tmp/telegram/`，上传 RustFS 后立即删除；网页经 Go 后端 JWT/Casbin 接口访问，不暴露 RustFS 凭据。

```bash
docker compose up -d --build
docker compose config
DATABASE_URL=postgresql://xianyu:xianyu_assistant@127.0.0.1:15432/xianyu_assistant python script/tel-bot/test_core.py
```

拒绝词只能拦截明显风险内容，不能判断版权授权；操作者仍须确认来源和保存权限。Office 文件当前提供下载，PDF、图片、文本可在线预览。
