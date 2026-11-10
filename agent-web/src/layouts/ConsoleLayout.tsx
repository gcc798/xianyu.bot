import {
  AppstoreOutlined,
  DatabaseOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
  SendOutlined,
  SettingOutlined,
  ShopOutlined,
} from "@ant-design/icons";
import { App, Button, Dropdown, Layout, Menu } from "antd";
import type { MenuProps } from "antd";
import { useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../app/AuthContext";
import { ADMIN_AUTHORITY_ID } from "../models/auth";
import styles from "./ConsoleLayout.module.css";

const { Header, Sider, Content } = Layout;

const routeTitles: Record<string, string> = {
  "/overview": "运营总览",
  "/bot": "Bot",
  "/resources": "资源库",
  "/resources/tasks": "采集任务",
  "/resources/files": "文件运维",
  "/profile": "个人设置",
  "/telegram/overview": "Telegram 总览",
  "/telegram/dialogs": "Telegram 会话",
  "/telegram/settings": "Telegram 设置",
  "/telegram/api-ids": "Telegram API ID",
  "/xianyu/accounts": "账号管理",
  "/xianyu/products": "商品管理",
  "/xianyu/orders": "订单管理",
  "/xianyu/messages": "消息记录",
  "/xianyu/reply-rules": "回复规则",
  "/xianyu/delivery-rules": "发货规则",
  "/xianyu/delivery-tasks": "发货任务",
  "/xianyu/events": "业务事件",
  "/system/users": "用户管理",
  "/system/dictionaries": "字典管理",
  "/system/login-logs": "登录日志",
  "/system/ai-settings": "AI 与 MCP",
};

export default function ConsoleLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { modal } = App.useApp();
  const { session, signOut } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const isAdmin = session?.user.authorityId === ADMIN_AUTHORITY_ID;
  const currentTitle = location.pathname.startsWith("/xianyu/products/")
    ? "商品详情"
    : location.pathname.startsWith("/xianyu/orders/")
      ? "订单详情"
      : routeTitles[location.pathname] || "控制台";
  const selectedPath = location.pathname.startsWith("/xianyu/products/")
    ? "/xianyu/products"
    : location.pathname.startsWith("/xianyu/orders/")
      ? "/xianyu/orders"
      : location.pathname;
  const menu: MenuProps["items"] = [
    { key: "/overview", icon: <AppstoreOutlined />, label: "总览" },
    { key: "/bot", icon: <RobotOutlined />, label: "Bot" },
    {
      key: "resources",
      icon: <DatabaseOutlined />,
      label: "资源中心",
      children: [
        { key: "/resources", label: "资源库" },
        { key: "/resources/tasks", label: "采集任务" },
        { key: "/resources/files", label: "文件运维" },
      ],
    },
    {
      key: "telegram",
      icon: <SendOutlined />,
      label: "Telegram",
      children: [
        { key: "/telegram/overview", label: "总览" },
        { key: "/telegram/dialogs", label: "会话" },
        { key: "/telegram/settings", label: "设置" },
        ...(isAdmin ? [{ key: "/telegram/api-ids", label: "API ID" }] : []),
      ],
    },
    {
      key: "xianyu",
      icon: <ShopOutlined />,
      label: "闲鱼自动化",
      children: [
        { key: "/xianyu/accounts", label: "账号管理" },
        { key: "/xianyu/products", label: "商品管理" },
        { key: "/xianyu/orders", label: "订单管理" },
        { key: "/xianyu/messages", label: "消息记录" },
        { key: "/xianyu/reply-rules", label: "回复规则" },
        { key: "/xianyu/delivery-rules", label: "发货规则" },
        { key: "/xianyu/delivery-tasks", label: "发货任务" },
        { key: "/xianyu/events", label: "业务事件" },
      ],
    },
    ...(isAdmin
      ? [
          {
            key: "system",
            icon: <SettingOutlined />,
            label: "系统设置",
            children: [
              { key: "/system/users", label: "用户管理" },
              { key: "/system/dictionaries", label: "字典管理" },
              { key: "/system/login-logs", label: "登录日志" },
              { key: "/system/ai-settings", label: "AI 与 MCP" },
            ],
          },
        ]
      : []),
  ];

  const confirmLogout = () => {
    modal.confirm({
      title: "退出当前账号？",
      content: "退出后需要重新登录才能继续使用控制台。",
      okText: "退出",
      cancelText: "取消",
      onOk: async () => {
        await signOut();
        navigate("/login", { replace: true });
      },
    });
  };

  return (
    <Layout className={`${styles.shell} ${collapsed ? styles.collapsed : ""}`} hasSider>
      <Sider
        width={228}
        collapsedWidth={64}
        collapsed={collapsed}
        collapsible
        trigger={null}
        theme="light"
        className={styles.sidebar}
      >
        <button className={styles.brand} title="返回总览" onClick={() => navigate("/overview")}>
          {collapsed ? "闲" : "闲鱼运营"}
        </button>
        <div className={styles.menuScroller}>
          <Menu
            mode="inline"
            inlineCollapsed={collapsed}
            items={menu}
            selectedKeys={[selectedPath]}
            onClick={({ key }) => navigate(key)}
            className={styles.menu}
          />
        </div>
        <div className={styles.collapseBar}>
          <Button
            type="text"
            block
            title={collapsed ? "展开菜单" : "收起菜单"}
            aria-label={collapsed ? "展开菜单" : "收起菜单"}
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed((value) => !value)}
          >
            {collapsed ? null : "收起菜单"}
          </Button>
        </div>
      </Sider>
      <Layout className={styles.workspace}>
        <Header className={styles.header}>
          <div className={styles.breadcrumb}>
            <span>
              {location.pathname.startsWith("/system")
                ? "系统设置"
                : location.pathname.startsWith("/telegram")
                  ? "Telegram"
                  : location.pathname.startsWith("/resources")
                    ? "资源中心"
                    : location.pathname === "/profile"
                      ? "账号"
                      : "闲鱼自动化"}
            </span>
            <i>/</i>
            <strong>{currentTitle}</strong>
          </div>
          <div className={styles.account}>
            <Dropdown
              menu={{
                items: [
                  { key: "profile", label: "个人设置" },
                  { type: "divider" },
                  { key: "logout", label: "退出登录", danger: true },
                ],
                onClick: ({ key }) => (key === "profile" ? navigate("/profile") : confirmLogout()),
              }}
              trigger={["click"]}
            >
              <Button type="text">{session?.user.nickName || session?.user.userName}</Button>
            </Dropdown>
          </div>
        </Header>
        <Content className={styles.content}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
