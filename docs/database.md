# 数据库

- 数据库：PostgreSQL 19，名称 `xianyu_bot`。
- 初始化入口：`server/internal/infra/migrations/init.sql`。
- 已有数据库的结构变更使用 Goose，迁移文件位于 `server/internal/infra/migrations/goose/`；Go 服务启动时不自动迁移。
- Telegram API ID 配置保存在 `s_telegram_api_configs`，API Hash 加密存储；变更配置前先备份数据库和凭据密钥。
- Compose 只在空数据卷执行初始化脚本。
- Go 和 Telegram 服务不动态建表。

执行未应用的迁移（在仓库根目录运行，使用已安装的 Goose CLI 和本机 PostgreSQL 映射端口）：

```bash
goose -dir server/internal/infra/migrations/goose postgres 'postgres://xianyu_bot:xianyu_bot@localhost:15432/xianyu_bot?sslmode=disable' up
```

表名前缀：

- `s_`：用户、字典和登录审计。
- `xy_`：闲鱼、Agent 和资源业务。
- `tel_`：Telegram 资源服务。

初始化数据只包含管理员、基础字典、Agent 默认配置和 Telegram 默认设置。`init.sql` 不负责升级已有数据库；需要保留数据时必须先提供增量迁移方案。

当前 Compose 使用官方 `postgres:19rc1-alpine`。PostgreSQL 19 正式版发布后改为 `postgres:19-alpine` 并重新执行空库验证。

重建空环境会删除数据库和对象存储数据：

```bash
docker compose down -v --remove-orphans
docker compose up -d --build
```
