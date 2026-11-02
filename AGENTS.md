# 仓库协作规范

本文件是 Codex 在本仓库中的唯一项目级规则入口。稳定工程规则写在这里；架构、领域和迁移事实写入 `docs/`。

## 当前架构

- `server/` 是 Go 后端，通用能力仅保留登录、管理员/普通用户、字典和登录审计。
- `server/internal/xianyu/` 承载闲鱼自动化和 Agent 业务，是正式领域模块，不是插件。
- `agent-web/` 是唯一维护的前端，使用 React 18 和 TypeScript。
- `web/` 是已废弃的 Gin-Vue-Admin/Vue 前端，只暂时用于行为对照、迁移验收和回滚。
- `server/internal/database/migrations/` 是 PostgreSQL 生产表结构的唯一事实源，服务启动时使用 Goose 执行迁移。
- 当前部署边界是可信内网；直接暴露公网属于范围变化，必须先补充网关、限流和认证设计。

修改模块边界或迁移范围前，先阅读 `docs/architecture.md`、`docs/project-scope.md` 和 `docs/migration/feature-matrix.md`。

## 前端归属

- 所有新页面、修复、依赖、测试和设计系统修改都放在 `agent-web/`。
- 不新增 JavaScript 或 JSX 源码，只使用 `.ts` 和 `.tsx`。
- 标准表单、表格、弹窗、分页、反馈和导航优先使用 Ant Design；Agent 和业务专属交互可以自研。
- 不再给 `web/` 增加功能、重构代码、升级依赖或重新生成构建产物。
- 可以读取 `web/` 核对旧行为和接口，但结果必须实现在 `agent-web/`。
- 对应 React 能力通过行为和权限验收后，才能删除 Vue 页面。
- 必须保留鉴权边界、加载/空/错误状态、键盘操作、响应式布局和危险操作确认。
- 重要 UI 修改必须在 `agent-web/` 执行 `npm run build`，并进行真实浏览器渲染检查。

## 后端边界

- API 层处理 HTTP 参数和响应，Service 层处理业务决策，`model/` 定义持久化与契约，`router/` 注册路由。
- Service 不依赖 `gin.Context`。
- 闲鱼依赖必须收敛在 `server/internal/xianyu/`；不得恢复 `server/plugin/xy` 或设计新的通用插件框架。
- 使用有领域语义的名称。除非职责确实准确，否则避免 `platform`、`runtime`、`manager`、`processor` 等宽泛命名。
- 复用系统登录、固定管理员边界、统一响应、字典和登录审计能力，不建立平行基础设施。
- 用户业务数据必须在 Service 层校验 `owner_id`。
- Cookie、API Key、Token、卡密等敏感值不得进入普通响应、日志、SSE、Agent Prompt、轨迹或浏览器状态。

## API 契约

- 兼容期内保持统一响应 `{ code, data, msg }` 和分页 `{ page, pageSize, total, list }`。
- Go 请求/响应结构与 TypeScript 类型的字段名和类型必须一致。
- 路径、方法、鉴权和响应变化都属于契约变化，必须同时修改所有调用方和相关文档。
- Swagger 等接口说明必须与真实行为一致。

## 数据库变更

- 所有生产表结构变化都必须新增 Goose 迁移，不得修改已经应用的基线迁移。
- 生产启动不得再使用 `AutoMigrate` 演进表结构。
- 只有完全由测试拥有的临时数据库可以在测试中使用 `AutoMigrate`。
- 迁移默认保留现有数据；有意破坏数据时必须明确记录并验证。
- 完成前检查迁移顺序、约束、索引和启动行为。

## 文档

- 架构和领域事实写入 `docs/`，不再创建 AI 专属记忆目录。
- 功能进入保留、待迁移、删除或范围外状态时，更新 `docs/project-scope.md`。
- 页面迁移验收后更新 `docs/migration/feature-matrix.md` 和 `docs/migration/vue-to-react.md`。
- 只记录已经验证的行为，并区分单测、集成测试、浏览器测试和真实账号测试。
- 遵守 `LICENSE` 和 `docs/upstream-and-licensing.md` 中的上游许可与署名边界。

## Skills

- 修改 Go、React、TypeScript 或 Vue 业务代码时，使用 `.agents/skills/idiomatic-business-code/`。
- 进行 UI 设计、审计、无障碍、响应式或视觉打磨时，使用 `.agents/skills/impeccable/`。
- 仓库 Skill 统一放在 `.agents/skills/<name>/SKILL.md`，不得在 `.codex`、`.claude`、`.cursor` 等目录复制规则。

## 验证

- 先运行最小相关测试，再运行受影响包或应用的完整检查。
- Go 修改必须格式化并测试受影响包。
- React 修改必须通过 TypeScript 生产构建。
- 数据库修改必须通过迁移测试。
- 只运行模拟或隔离测试时，不得声称真实外部联调成功。
