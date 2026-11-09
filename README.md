# xianyu.bot

面向闲鱼自动化和可审计 Agent 的 Go 模块化单体。

## 目录

```text
server/        Go API、闲鱼业务、Agent 和 PostgreSQL
agent-web/     React 19 + TypeScript 前端
telegram.bot/ Telegram 资源与 MCP 服务
docs/          当前架构和工程规则
```

工程文档见 [docs/README.md](docs/README.md)。

## 启动

```bash
docker compose up -d --build
```

默认地址：

- Web：<http://127.0.0.1:18080/agent/>
- API：<http://127.0.0.1:18888>
- Telegram：<http://127.0.0.1:18081>
- PostgreSQL：`127.0.0.1:15432`

初始管理员为 `admin / 123456`，首次登录后必须修改密码。生产环境必须设置 `POSTGRES_PASSWORD` 和 `XIANYU_JWT_SIGNING_KEY`。

## 验证

```bash
cd server && go test ./... && go vet ./...
cd ../agent-web && npm run build
docker compose config
```

数据库初始化见 [docs/database.md](docs/database.md)，前端现状见 [docs/frontend.md](docs/frontend.md)，许可边界见 [docs/third-party-licensing.md](docs/third-party-licensing.md)。
