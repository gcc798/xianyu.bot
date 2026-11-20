# 数据库

- 数据库：PostgreSQL 19，名称 `xianyu_bot`。
- 初始化基线：`server/internal/infra/migrations/init.sql`，仅用于空库，包含到 Goose 迁移 003 为止的结构与基础数据；基线版本号固定为 003。
- Go 服务启动时，空库先执行初始化基线，再自动应用 Goose 004 及之后未执行的迁移；已有数据库只按 Goose 版本升级，绝不重放 `init.sql`。非空但无 Goose 历史的数据库会停止启动，需先人工核对并建立迁移基线。
- **数据库结构迭代只新增 Goose 迁移，禁止把后续表、字段、索引或数据修订写回 `init.sql`。** `init.sql` 是不可变的空库基线，不是 schema 的持续维护入口。
- `goose/` 单独存放可按版本执行的增量 SQL；完整初始化脚本不放入该目录，避免被当作普通迁移重复回放。
- 每次数据库变更都要新增带唯一版本号的 Goose SQL，同时编写 `Up` 和可逆的 `Down`；不得用 `AutoMigrate` 或手工 SQL 作为部署步骤。新增迁移须覆盖两种验证：已有库从当前版本升级，以及空库从固定基线应用全部迁移。
- 迁移 SQL 遵循仓库约束：不新增物理外键、`CHECK`、`ENUM`、自定义 `DOMAIN` 或触发器；可按需使用主键、唯一约束和索引，业务取值校验放在应用层。
- 新增 MCP 工具目录和执行治理字段的迁移为 `202611200009_mcp_tool_catalog.sql`；该迁移会补齐已运行数据库缺少的 `xy_mcp_tools` 及相关列。
- 字典表使用 `s_dict` 和 `s_dict_detail`；增量迁移 006 将早期表名原位重命名，保留字典数据与字段。
- 增量迁移 007 删除初始化基线中的旧 JWT 黑名单表；退出登录只清除客户端令牌，服务端仍校验签名和过期时间。
- Telegram API ID 配置保存在 `s_telegram_api_configs`，API Hash 加密存储；变更配置前先备份数据库和凭据密钥。
- Compose 只在空数据卷执行初始化脚本。
- Go 和 Telegram 服务不动态建表。

可手动查看迁移状态或执行未应用的迁移（在仓库根目录运行，使用已安装的 Goose CLI 和本机 PostgreSQL 映射端口）：

```bash
goose -dir server/internal/infra/migrations/goose postgres 'postgres://xianyu_bot:xianyu_bot@localhost:15432/xianyu_bot?sslmode=disable' status
goose -dir server/internal/infra/migrations/goose postgres 'postgres://xianyu_bot:xianyu_bot@localhost:15432/xianyu_bot?sslmode=disable' up
```

生产升级前先备份并确认目标数据库；正常部署由 Go 服务启动时应用迁移。不要在已有数据库上执行 `init.sql`，不要通过删除数据卷来绕过迁移问题。

表名前缀：

- `s_`：用户、字典和登录审计。
- `xy_`：闲鱼、Agent 和资源业务。
- `tel_`：Telegram 资源服务。

初始化数据只包含管理员、基础字典、Agent 默认配置和 Telegram 默认设置。服务启动不会在已有数据上重放 `init.sql`。

当前 Compose 使用官方 `postgres:19rc1-alpine`。PostgreSQL 19 正式版发布后改为 `postgres:19-alpine` 并重新执行空库验证。

重建空环境会删除数据库和对象存储数据：

```bash
docker compose down -v --remove-orphans
docker compose up -d --build
```
