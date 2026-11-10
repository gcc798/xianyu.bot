# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

主要用户是可信内网中的闲鱼运营人员和系统管理员。他们在桌面浏览器中管理闲鱼账号、商品、订单、消息、自动化规则、Bot 会话和外部资源任务。

管理员还负责系统用户、业务字典、登录审计、AI 模型和 MCP Server 配置；普通用户只能使用其权限范围内的业务能力。

## Product Purpose

产品把闲鱼运营、自动回复、自动发货、自动续售、Agent 协作和 Telegram 资源能力收束到同一个 React 运营控制台。成功意味着用户能够在清晰的权限边界内完成日常运营，并能检查关键业务状态、失败原因和 Agent 轨迹。

## Positioning

控制台不只是展示闲鱼数据，而是把闲鱼协议连接、确定性业务自动化和可审批的 Agent Tool 调用统一到同一套账号归属、审计和资源交付链路中。

## Operating Context

- 用户以商品、订单、消息和规则列表完成高频运营任务，并通过详情页检查关联状态和事件。
- Bot 页面用于发起运营问题或任务，并在同一会话中查看对话、审批、工具调用和轨迹。
- Telegram 提供外部资源检索渠道，RustFS 保存交付资源。
- 资源中心统一呈现可供闲鱼上架的资源、Telegram 批量采集任务和文件运维；资源及采集结果按用户隔离。
- 系统部署在可信内网，不以公网 SaaS 为当前目标。

## Capabilities and Constraints

- `agent-web/` 是唯一维护和部署的 Web 前端，使用 React 19、TypeScript、Vite、React Router 和 Ant Design。
- 前端使用当前唯一的 Go API 合同，不保留废弃字段、路由或双响应格式的兼容代码。
- 保留管理员与普通用户边界、用户业务数据所有权、加载/空/错误状态和危险操作确认。
- 正式页面不使用模拟数据，不以全量列表拼接替代缺失的详情接口。
- Cookie、Token、API Key、卡密和 MCP 凭据不得进入普通响应、日志、SSE、Prompt 或轨迹；管理员模型配置接口可返回完整 API Key 供核对和编辑。
- Agent 不提供会话级模型选择；已启用且支持 Tool 的模型按优先级从小到大尝试，调用失败时无感降级，并在轨迹记录切换过程。
- Agent 新会话在首条消息后以第一句话命名，用户可以随时修改；侧栏只省略显示，不截断已保存名称。
- 当前只交付桌面浏览器界面；浏览器缩放后的有效宽度仍需保持可用，不单独设计移动端页面。

## Brand Commitments

- 产品名称使用“闲鱼运营”。
- 一级菜单使用“Bot”和“闲鱼自动化”。
- 界面语言为简体中文，语气直接、克制，不添加解释性营销文案。
- 视觉采用明亮中性的企业级运营控制台方向。
- 字体、对话与轨迹的信息密度参考 DeepSeek Harness，但不引入工作区概念。

## Evidence on Hand

- 真实业务和 API 合同位于 `server/internal/api/router/`、`server/internal/model/request/` 和 `server/internal/model/response/`。
- 正式页面和交互实现位于 `agent-web/src/pages/`，接口调用和 DTO 分别位于 `agent-web/src/api/`、`agent-web/src/models/`。
- 仓库没有可公开使用的真实账号、凭据、交易数据或品牌图片；不得自行编造真实验收结果。

## Product Principles

1. 业务状态和下一步操作优先于装饰与说明文案。
2. 一个全局 Layout、一套组件语言和一份权限边界覆盖全部页面。
3. 列表保持高信息密度，详情按数据领域分区并保持横平竖直。
4. Agent 行为必须可检查、可审批、可取消，危险操作必须明确确认。
5. 前后端合同不一致时修正唯一合同，不在前端增加兼容分支。

## Accessibility & Inclusion

正式页面必须支持键盘焦点、语义化表单标签、可辨识的错误状态和不依赖颜色单独传达的业务状态。
