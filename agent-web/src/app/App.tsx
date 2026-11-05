import { App as AntApp, ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import { RouterProvider } from "react-router-dom";
import { AuthProvider } from "./AuthProvider";
import { router } from "../router";

export default function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      componentSize="medium"
      theme={{
        token: {
          colorPrimary: "#2f67e8",
          colorInfo: "#2f67e8",
          colorSuccess: "#16875f",
          colorWarning: "#b96816",
          colorError: "#c93c37",
          colorText: "#20242b",
          colorTextSecondary: "#67707e",
          colorBorder: "#dce1e8",
          colorBgLayout: "#f4f6f8",
          borderRadius: 6,
          controlHeight: 36,
          fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
          fontSize: 14,
          boxShadowSecondary: "0 10px 32px rgba(18, 31, 53, 0.12)",
        },
        components: {
          Layout: { bodyBg: "#f4f6f8", headerBg: "#ffffff", siderBg: "#f7f8fa" },
          Menu: {
            itemBg: "transparent",
            itemSelectedBg: "#e8eefb",
            itemSelectedColor: "#245fd6",
            itemHoverBg: "#eef1f5",
            itemBorderRadius: 6,
          },
          Table: {
            headerBg: "#ffffff",
            headerColor: "#20242b",
            rowHoverBg: "#f6f8fb",
            borderColor: "#e1e5eb",
            cellPaddingBlockSM: 8,
          },
          Button: { borderRadius: 5, primaryShadow: "none" },
          Input: { activeShadow: "0 0 0 2px rgba(47, 103, 232, 0.12)" },
          Select: { activeOutlineColor: "rgba(47, 103, 232, 0.12)" },
          Modal: { borderRadiusLG: 8 },
        },
      }}
    >
      <AntApp>
        <AuthProvider>
          <RouterProvider router={router} />
        </AuthProvider>
      </AntApp>
    </ConfigProvider>
  );
}
