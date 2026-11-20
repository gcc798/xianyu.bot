<div align="center">
  <img src="docs/xianyu-bot-logo.svg" width="560" alt="xianyu.bot" />
  <p><strong>面向闲鱼运营的可审计 Agent 控制台</strong></p>
  <p>
    <img src="https://img.shields.io/badge/Go-1.26.10-00ADD8?logo=go&logoColor=white" alt="Go 1.26.10" />
    <img src="https://img.shields.io/badge/React-19.3.1-149ECA?logo=react&logoColor=white" alt="React 19.3.1" />
    <img src="https://img.shields.io/badge/TypeScript-5.9.3-3178C6?logo=typescript&logoColor=white" alt="TypeScript 5.9.3" />
    <img src="https://img.shields.io/badge/Ant_Design-6.6.8-1677FF?logo=antdesign&logoColor=white" alt="Ant Design 6.6.8" />
  </p>
  <p>
    <img src="https://img.shields.io/badge/Eino-0.9.22-7C3AED" alt="Eino 0.9.22" />
    <img src="https://img.shields.io/badge/Python-3.12-3776AB?logo=python&logoColor=white" alt="Python 3.12" />
    <img src="https://img.shields.io/badge/PostgreSQL-19_RC1-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL 19 RC1" />
    <img src="https://img.shields.io/badge/RustFS-latest-CE422B?logo=rust&logoColor=white" alt="RustFS latest" />
    <img src="https://img.shields.io/badge/License-Apache_2.0-4C1?logo=apache&logoColor=white" alt="Apache License 2.0" />
  </p>
</div>

xianyu.bot 将闲鱼商品与消息自动化、Agent 会话与审批、MCP 工具调用，以及 Telegram 资源交付集中到一个控制台，让每次自动化操作都可确认、可追踪、可审计。

## 核心能力

| 能力 |
| --- |
| 🤖 **Agent 工作台**：对话执行、敏感操作确认、完整调用轨迹 |
| 🛍️ **闲鱼自动化**：账号、商品、订单、回复与发货规则 |
| 🔌 **MCP 与 Tool**：统一配置、权限过滤、健康检查与按需调用 |
| 📦 **资源交付**：资源检索、对象存储与 Telegram 文件交付 |

## 界面预览

<a href="docs/agent-approval.png">
  <img src="docs/agent-approval.png" alt="Agent 对话与敏感操作审批" />
</a>
<p align="center"><sub>Agent 对话 · 敏感操作需要确认后执行</sub></p>

<a href="docs/agent-trajectory.png">
  <img src="docs/agent-trajectory.png" alt="Agent 完整执行轨迹" />
</a>
<p align="center"><sub>执行轨迹 · 模型、工具与审批过程全程可追溯</sub></p>

## Docker Compose

```bash
docker compose up -d --build
```

默认管理员为 `admin / 123456`，首次登录后请立即修改密码。

## License

[Apache License 2.0](LICENSE)
