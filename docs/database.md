# 数据库

- 数据库：PostgreSQL 19，名称 `xianyu_bot`。
- 初始化入口：`server/internal/infra/migrations/init.sql`。
- Compose 只在空数据卷执行初始化脚本。
- Go 和 Telegram 服务不动态建表。

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
