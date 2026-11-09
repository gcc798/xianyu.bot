# 前端

`agent-web/` 是唯一维护的前端，使用 React 19、TypeScript、Vite、React Router 和 Ant Design。当前包含：

- 登录与退出；
- 运营总览；
- 闲鱼账号、商品、订单、消息、回复规则、发货规则、发货任务和业务事件；
- 管理员模型、MCP Server 和 Agent 运行边界配置；
- 统一资源列表、来源详情和下载；
- Telegram 状态、扫码登录、扫描、会话、资源任务、资源文件和设置管理；
- 用户、字典和登录日志；
- Bot 对话、审批、轨迹和工具检查。

正式页面通过 Browser Router 挂载在 `/agent/`，使用同一控制台 Layout。页面按路由懒加载，管理员页面同时受菜单、路由守卫和后端权限保护。
