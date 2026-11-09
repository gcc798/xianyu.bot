import { Navigate, createBrowserRouter } from "react-router-dom";
import ConsoleLayout from "../layouts/ConsoleLayout";
import RouteErrorPage, { NotFoundPage } from "../pages/error/RouteErrorPage";
import { AdminRoutes, ProtectedRoutes, RouteFallback } from "./RouteGuards";

export const router = createBrowserRouter(
  [
    {
      path: "/login",
      lazy: () => import("../pages/login/LoginPage"),
      HydrateFallback: RouteFallback,
      errorElement: <RouteErrorPage />,
    },
    {
      element: <ProtectedRoutes />,
      errorElement: <RouteErrorPage />,
      children: [
        {
          element: <ConsoleLayout />,
          children: [
            { index: true, element: <Navigate to="/overview" replace /> },
            {
              path: "overview",
              lazy: () => import("../pages/dashboard/DashboardPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "bot",
              lazy: () => import("../pages/bot/BotPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "resources",
              lazy: () => import("../pages/resources/ResourcesPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "profile",
              lazy: () => import("../pages/profile/ProfilePage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "telegram/overview",
              lazy: () => import("../pages/telegram/overview/TelegramOverviewPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "telegram/dialogs",
              lazy: () => import("../pages/telegram/dialogs/TelegramDialogsPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "telegram/resource-tasks",
              lazy: () => import("../pages/telegram/resource-tasks/TelegramResourceTasksPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "telegram/resource-files",
              lazy: () => import("../pages/telegram/resource-files/TelegramResourceFilesPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "telegram/settings",
              lazy: () => import("../pages/telegram/settings/TelegramSettingsPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/accounts",
              lazy: () => import("../pages/xianyu/accounts/AccountsPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/products",
              lazy: () => import("../pages/xianyu/products/ProductsPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/products/:id",
              lazy: () => import("../pages/xianyu/products/ProductDetailPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/orders",
              lazy: () => import("../pages/xianyu/orders/OrdersPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/orders/:id",
              lazy: () => import("../pages/xianyu/orders/OrderDetailPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/messages",
              lazy: () => import("../pages/xianyu/messages/MessagesPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/reply-rules",
              lazy: () => import("../pages/xianyu/reply-rules/ReplyRulesPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/delivery-rules",
              lazy: () => import("../pages/xianyu/delivery-rules/DeliveryRulesPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/delivery-tasks",
              lazy: () => import("../pages/xianyu/delivery-tasks/DeliveryTasksPage"),
              HydrateFallback: RouteFallback,
            },
            {
              path: "xianyu/events",
              lazy: () => import("../pages/xianyu/events/BusinessEventsPage"),
              HydrateFallback: RouteFallback,
            },
            {
              element: <AdminRoutes />,
              children: [
                {
                  path: "system/users",
                  lazy: () => import("../pages/system/users/UsersPage"),
                  HydrateFallback: RouteFallback,
                },
                {
                  path: "system/dictionaries",
                  lazy: () => import("../pages/system/dictionaries/DictionariesPage"),
                  HydrateFallback: RouteFallback,
                },
                {
                  path: "system/login-logs",
                  lazy: () => import("../pages/system/login-logs/LoginLogsPage"),
                  HydrateFallback: RouteFallback,
                },
                {
                  path: "system/ai-settings",
                  lazy: () => import("../pages/system/ai-settings/AISettingsPage"),
                  HydrateFallback: RouteFallback,
                },
              ],
            },
          ],
        },
      ],
    },
    { path: "*", element: <NotFoundPage /> },
  ],
  { basename: "/agent" },
);
