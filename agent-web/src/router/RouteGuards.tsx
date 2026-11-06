import { Spin } from "antd";
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../app/AuthContext";
import { ADMIN_AUTHORITY_ID } from "../models/auth";

export function RouteFallback() {
  return (
    <div className="app-loading">
      <Spin size="large" />
    </div>
  );
}

/** 等待会话校验完成后再决定跳转，避免刷新页面时短暂显示登录页。 */
export function ProtectedRoutes() {
  const { session, isReady } = useAuth();
  if (!isReady) return <RouteFallback />;
  if (!session) return <Navigate to="/login" replace />;
  return <Outlet />;
}

/** 管理菜单之外再增加路由级拦截；最终权限仍由后端校验。 */
export function AdminRoutes() {
  const { session } = useAuth();
  if (session?.user.authorityId !== ADMIN_AUTHORITY_ID) return <Navigate to="/bot" replace />;
  return <Outlet />;
}
