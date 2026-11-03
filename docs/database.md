# PostgreSQL 初始化

## 当前基线

PostgreSQL 是唯一应用数据库。表结构与最小基础数据只有一个入口：

```text
server/internal/infra/migrations/init.sql
```

Docker Compose 将该文件挂载到 `/docker-entrypoint-initdb.d/001-init.sql`。PostgreSQL 仅在数据目录为空时执行它；Go 和 Telegram 服务都不会运行 DDL 或自动迁移。

初始化数据只包含：

- 超级管理员 `admin`，初始密码 `123456`；
- 管理员、普通用户和用户状态字典；
- Agent 默认运行上限；
- Telegram 默认设置。

通用基础表使用 `s_` 前缀，闲鱼业务表使用 `xy_`，Telegram 表使用 `tel_`。

首次登录后必须修改管理员密码。仓库不得提交真实用户、邮箱凭据、Token、Cookie、会话、资源或业务数据。

## 版本与配置

当前 Compose 使用官方 `postgres:19rc1-alpine`。PostgreSQL 19 尚未正式发布，RC 镜像只适合开发和验收；正式版发布后切换到 `postgres:19-alpine` 并重新执行空库验证。

数据库名和用户均为 `xianyu_bot`。密码通过 `POSTGRES_PASSWORD` 注入，后端通过 `XIANYU_POSTGRES_PASSWORD` 接收同一值；JWT 密钥通过 `XIANYU_JWT_SIGNING_KEY` 注入。

## 变更规则

- `init.sql` 只面向空数据库，不承担已有实例升级。
- 当前阶段不得新增其他建表脚本，也不得在应用启动时执行 `CREATE TABLE` 或 `AutoMigrate`。
- 只有完全由测试拥有的临时数据库可以使用 `AutoMigrate`。
- 修改表结构后必须从空数据卷启动，检查表、约束、索引、序列、基础数据、登录和鉴权 API。
- 需要保留已有部署数据时，先设计并引入增量迁移机制，不能直接重建数据卷。

## 本地启动

```bash
docker compose down -v --remove-orphans
docker compose up -d --build
```

`down -v` 会删除当前 Compose 项目的 PostgreSQL 和 RustFS 数据卷，只能用于明确需要全新环境的场景。
