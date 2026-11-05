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
            { path: "xianyu/accounts", lazy: () => import("../pages/xianyu/accounts/AccountsPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/products", lazy: () => import("../pages/xianyu/products/ProductsPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/orders", lazy: () => import("../pages/xianyu/orders/OrdersPage"), HydrateFallback: RouteFallback },
            { path: "xianyu/messages", lazy: () => import("../pages/xianyu/messages/MessagesPage"), HydrateFallback: RouteFallback },
            {
              element: <AdminRoutes />,
              children: [
                { path: "system/users", lazy: () => import("../pages/system/users/UsersPage"), HydrateFallback: RouteFallback },
                { path: "system/dictionaries", lazy: () => import("../pages/system/dictionaries/DictionariesPage"), HydrateFallback: RouteFallback },
                { path: "system/login-logs", lazy: () => import("../pages/system/login-logs/LoginLogsPage"), HydrateFallback: RouteFallback },
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
