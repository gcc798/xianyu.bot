# 仓库规则入口

本文件只负责规则导航，不承载具体的业务、技术或交付规范。

## 规则加载

开始工作前，根据任务范围完整阅读对应规范；同时涉及多个范围时，必须读取全部相关文件：

- 任何软件开发、工程配置、文档或交付任务：`.specs/project-development.md`
- 涉及 Go 代码的新增、修改或评审：`.specs/go-code-style.md`
- 涉及后端分层、API、鉴权、安全、数据模型或数据库：`.specs/backend-development.md`
- 涉及 React、TypeScript、前端交互或页面样式：`.specs/frontend-development.md`

规则文件中的要求仅适用于其声明的范围，不得机械套用到无关任务。当前用户要求优先于仓库规范；规则冲突时，更具体的规则优先。

## Skills 加载

- 修改 Go、React 或 TypeScript 业务代码时，使用 `.agents/skills/idiomatic-business-code/`。
- 进行 UI 设计、审计或视觉调整时，使用 `.agents/skills/impeccable/`。
