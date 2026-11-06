import { createContext, useContext } from "react";
import type { AuthSession, LoginRequest } from "../models/auth";

export type AuthContextValue = {
  session: AuthSession | null;
  isReady: boolean;
  signIn: (input: LoginRequest) => Promise<void>;
  signOut: () => Promise<void>;
  refreshUser: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

/** 读取当前登录会话；只能在 AuthProvider 管理的组件树中调用。 */
export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth 必须在 AuthProvider 内使用");
  return value;
}
