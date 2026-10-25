# Telegram Analyzer & Cleaner

本地、单用户的 Telegram 群聊/频道分析与手工清理工具。只读取已加入的 dialogs 元数据；唯一写入 Telegram 的操作是经二次确认后退出所选群聊/频道。

## 安装与启动

1. 在 [my.telegram.org](https://my.telegram.org) 创建应用，获取 API ID 与 API Hash。
2. 配置并启动：

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# 编辑 .env
python3 app.py
```

如果当前网络需要 SOCKS5 代理，在 `.env` 中配置：

```env
TG_PROXY_HOST=127.0.0.1
TG_PROXY_PORT=17891
```

浏览器打开 <http://127.0.0.1:8080>。首次使用点击“生成二维码”，用手机 Telegram 的“设置 → 设备 → 链接桌面设备”扫码。Session、SQLite 数据库保存在 `data/`，不要分享该目录。

## 使用

- 快速扫描只遍历 dialogs，保存名称、类型、人数、最新消息时间、未读、静音、归档和创建者状态。
- 不读取消息历史，不下载视频、图片或文件；人数缺失时显示 `-`，不额外请求成员列表。
- 群聊页一次展示全部结果，集成名称模糊搜索、筛选、升降序、Telegram 跳转、勾选退出和 CSV 导出。
- 退出前会再弹窗确认；每项结果写入 `leave_history`。

运行最小自检：

```bash
python test_core.py
```

局域网访问会暴露管理页面；仅在可信网络使用，并用 macOS 防火墙限制访问。Bootstrap CSS 使用 CDN，离线时功能仍可用但样式会变简陋。
