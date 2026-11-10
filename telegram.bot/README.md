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

浏览器打开 <http://127.0.0.1:8080>。首次使用点击“生成二维码”，用手机 Telegram 的“设置 → 设备 → 链接桌面设备”扫码。登录 session 保存在 PostgreSQL `xianyu_bot.tel_settings` 中，应按凭据保护，不要导出或分享。

Streamable HTTP MCP 端点为 `/mcp/`，固定提供渠道、来源、资源和通用任务工具。服务部署在可信内网中，MCP 与 Go 内部接口不额外校验共享令牌。资源任务和文件操作仍按 `owner_id` 做用户隔离。

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

Compose 中的服务统一使用 `xianyu.bot` 网络。闲鱼助手后台映射到宿主机 `18080`，PostgreSQL 映射到 `15432`，RustFS 的 S3 API/控制台分别映射到 `19000`/`19001`，Telegram Web UI 映射到 `127.0.0.1:18081`。

Telegram 与 Go 后端共用 `xianyu_bot` 数据库，Telegram 表统一使用 `tel_` 前缀；PostgreSQL 通过宿主机 `15432` 端口连接。表结构由根工程的单一 `init.sql` 初始化，Telegram 进程不动态建表。

Telegram 应用凭据由 Go 维护在 `s_telegram_api_configs`，API Hash 加密存储。Compose 项目根目录 `.env` 默认使用数据库配置：`TG_API_CREDENTIAL_SOURCE=database`；设为 `env` 时沿用 `telegram.bot/.env` 中的 `TG_API_ID` 和 `TG_API_HASH`，不会在两种来源间自动回退。数据库模式下 Python 每个资源任务开始前选择未限流且 API ID 最大的配置；任务中遇到 FloodWait 会记录限流截止时间、切换下一个可用配置并重试当前 Telegram 操作，累计 5 个不同 API ID 限流后任务失败。切换在同一登录 Session 上进行；若 Telegram 未授权新客户端，则任务明确失败。管理列表不会返回明文 API Hash。

资源任务表为 `tel_resource_tasks`、`tel_resource_files`、`tel_resource_sources`。任务和文件记录带有所有者 ID；网页请求经 Go JWT 代理传递已认证用户 ID，Python 查询、预览、下载和删除均按所有者过滤。任务与消息来源记录用于 Telegram 重试和运维；文件下载完成后由 Go 在 `xy_resources` 统一登记，可供资源库和闲鱼上架流程使用。文件临时落在项目根目录 `tmp/telegram/`，上传 RustFS 后立即删除；网页不暴露 RustFS 凭据。

现有数据库升级时必须先备份，并将没有所有者字段的历史 Telegram 任务和文件显式归属给管理员（用户 ID 1）；禁止将历史共享数据直接开放给普通用户。初始化 SQL 仅适用于空库，不会自动升级已有数据库。

```bash
docker compose up -d --build
docker compose config
DATABASE_URL=postgresql://xianyu_bot:xianyu_bot@127.0.0.1:15432/xianyu_bot python telegram.bot/test_core.py
```

拒绝词只能拦截明显风险内容，不能判断版权授权；操作者仍须确认来源和保存权限。资源类型由服务端文件字典统一维护：PDF、Markdown、文本、CSV、JSON、XML、JPEG、PNG、WebP、MP4、WebM、MOV、M4V、MP3、M4A、WAV、OGG 支持在线预览；Office、OpenDocument、EPUB、MOBI、RTF 和 ZIP 仅提供下载。
