# PostgreSQL 与 Goose

## 当前状态

PostgreSQL 是重构后的应用数据库。现有应用数据库复制出的表结构已经保存为 Goose 基线 `202611010001`：

```text
server/internal/database/migrations/202611010001_baseline.sql
```

服务启动时通过 `server/internal/database.Migrate` 执行待应用迁移。

## 基线认领

空 PostgreSQL 数据库由 Goose 执行基线。

已有复制数据库没有 Goose 历史时，只有以下三个代表表同时存在才允许认领基线：

- `sys_users`；
- `xy_accounts`；
- `xy_agent_sessions`。

只有部分表匹配时启动失败，不猜测数据库状态。

## 后续规则

- 不得修改已经应用的基线。
- 每次生产表结构变化新增时间戳 `*.sql` 迁移。
- 迁移必须确定、可审查并面向 PostgreSQL。
- 默认保留数据；有意破坏时明确记录并验证。
- 迁移评审同时检查约束和索引。
- 同时测试空库迁移和已有基线认领。
- 生产启动不得使用 `AutoMigrate`。
- 测试只有在完全拥有临时表结构时才能使用 `AutoMigrate`。

## 数据验证

复制或转换数据库时：

1. 记录源库和目标库标识；
2. 比较保留表和行数；
3. 检查主键、唯一约束、外键和序列；
4. 验证加密凭据仍不可读且可用；
5. 执行登录、鉴权、闲鱼列表和 Agent 会话冒烟；
6. 验收前保留可恢复的源库备份。

`tmp/` 下的 dump 是本地恢复产物，不属于版本化迁移历史。
