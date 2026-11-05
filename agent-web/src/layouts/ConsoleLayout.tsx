import { App, Button, Layout, Menu } from "antd";
import type { MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../app/AuthProvider";
import { ADMIN_AUTHORITY_ID } from "../models/auth";
import styles from "./ConsoleLayout.module.css";

const { Header, Sider, Content } = Layout;

const routeTitles: Record<string, string> = {
  "/bot": "Bot",
  "/system/users": "用户管理",
  "/system/dictionaries": "字典管理",
  "/system/login-logs": "登录日志",
};

export default function ConsoleLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { modal } = App.useApp();
  const { session, signOut } = useAuth();
  const isAdmin = session?.user.authorityId === ADMIN_AUTHORITY_ID;
  const menu: MenuProps["items"] = [
    { key: "/bot", label: "Bot" },
    ...(isAdmin
      ? [{
          key: "system",
          label: "系统设置",
          children: [
            { key: "/system/users", label: "用户管理" },
            { key: "/system/dictionaries", label: "字典管理" },
            { key: "/system/login-logs", label: "登录日志" },
          ],
        }]
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
    <Layout className={styles.shell} hasSider>
      <Sider width={228} theme="light" className={styles.sidebar}>
        <button className={styles.brand} onClick={() => navigate("/bot")}>闲鱼运营</button>
        <Menu
          mode="inline"
          items={menu}
          selectedKeys={[location.pathname]}
          defaultOpenKeys={["system"]}
          onClick={({ key }) => navigate(key)}
          className={styles.menu}
        />
      </Sider>
      <Layout className={styles.workspace}>
        <Header className={styles.header}>
          <div className={styles.breadcrumb}>
            <span>{location.pathname.startsWith("/system") ? "系统设置" : "闲鱼自动化"}</span>
            <i>/</i>
            <strong>{routeTitles[location.pathname] || "控制台"}</strong>
          </div>
          <div className={styles.account}>
            <span>{session?.user.nickName || session?.user.userName}</span>
            <Button type="text" onClick={confirmLogout}>退出</Button>
          </div>
        </Header>
        <Content className={styles.content}><Outlet /></Content>
      </Layout>
    </Layout>
  );
}
