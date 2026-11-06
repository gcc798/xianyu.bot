import { Spin } from "antd";
import { Navigate, Outlet, createBrowserRouter } from "react-router-dom";
import { useAuth } from "../app/AuthProvider";
import ConsoleLayout from "../layouts/ConsoleLayout";

function RouteFallback() {
  return <div className="app-loading"><Spin size="large" /></div>;
}

function ProtectedRoutes() {
  const { session, isReady } = useAuth();
  if (!isReady) return <div className="app-loading"><Spin size="large" /></div>;
  if (!session) return <Navigate to="/login" replace />;
  return <Outlet />;
}

function AdminRoutes() {
  const { session } = useAuth();
  if (session?.user.authorityId !== 888) return <Navigate to="/bot" replace />;
  return <Outlet />;
}

export const router = createBrowserRouter(
  [
    { path: "/login", lazy: () => import("../pages/login/LoginPage"), HydrateFallback: RouteFallback },
    {
      element: <ProtectedRoutes />,
      children: [
        {
          element: <ConsoleLayout />,
          children: [
            { index: true, element: <Navigate to="/bot" replace /> },
            { path: "bot", lazy: () => import("../pages/bot/BotPage"), HydrateFallback: RouteFallback },
            { path: "resources", lazy: () => import("../pages/resources/ResourcesPage"), HydrateFallback: RouteFallback },
            { path: "telegram/overview", lazy: () => import("../pages/telegram/overview/TelegramOverviewPage"), HydrateFallback: RouteFallback },
            { path: "telegram/dialogs", lazy: () => import("../pages/telegram/dialogs/TelegramDialogsPage"), HydrateFallback: RouteFallback },
            { path: "telegram/settings", lazy: () => import("../pages/telegram/settings/TelegramSettingsPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/accounts", lazy: () => import("../pages/xianyu/accounts/AccountsPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/products", lazy: () => import("../pages/xianyu/products/ProductsPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/orders", lazy: () => import("../pages/xianyu/orders/OrdersPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/messages", lazy: () => import("../pages/xianyu/messages/MessagesPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/reply-rules", lazy: () => import("../pages/xianyu/reply-rules/ReplyRulesPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/delivery-rules", lazy: () => import("../pages/xianyu/delivery-rules/DeliveryRulesPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/delivery-tasks", lazy: () => import("../pages/xianyu/delivery-tasks/DeliveryTasksPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/events", lazy: () => import("../pages/xianyu/events/BusinessEventsPage"), HydrateFallback: RouteFallback },
            {
              element: <AdminRoutes />,
              children: [
                { path: "system/users", lazy: () => import("../pages/system/users/UsersPage"), HydrateFallback: RouteFallback },
                { path: "system/dictionaries", lazy: () => import("../pages/system/dictionaries/DictionariesPage"), HydrateFallback: RouteFallback },
                { path: "system/login-logs", lazy: () => import("../pages/system/login-logs/LoginLogsPage"), HydrateFallback: RouteFallback },
                { path: "system/ai-settings", lazy: () => import("../pages/system/ai-settings/AISettingsPage"), HydrateFallback: RouteFallback },
              ],
            },
          ],
        },
      ],
    },
    { path: "*", element: <Navigate to="/" replace /> },
  ],
  { basename: "/agent" },
);
