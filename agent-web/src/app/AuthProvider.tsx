import { useEffect, useMemo, useRef, useState } from "react";
import type { PropsWithChildren } from "react";
import { getCurrentUser, login, logout, readAuthSession, storeAuthSession } from "../api/auth";
import type { AuthSession } from "../models/auth";
import { AuthContext } from "./AuthContext";
import type { AuthContextValue } from "./AuthContext";

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession | null>(() => readAuthSession());
  const initialSession = useRef(session).current;
  const [isReady, setIsReady] = useState(initialSession === null);

  // 启动时只校验持久化会话；后续登录和资料刷新由对应动作直接更新状态。
  useEffect(() => {
    if (!initialSession) return;
    let active = true;
    void getCurrentUser()
      .then(({ userInfo }) => {
        if (!active) return;
        const renewed = readAuthSession();
        const current = { ...(renewed || initialSession), user: userInfo };
        storeAuthSession(current);
        setSession(current);
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setIsReady(true);
      });
    return () => {
      active = false;
    };
  }, [initialSession]);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isReady,
      signIn: async (input) => setSession(await login(input)),
      signOut: async () => {
        await logout();
        setSession(null);
      },
      refreshUser: async () => {
        const { userInfo } = await getCurrentUser();
        setSession((current) => {
          if (!current) return current;
          const renewed = readAuthSession();
          const next = { ...(renewed || current), user: userInfo };
          storeAuthSession(next);
          return next;
        });
      },
    }),
    [isReady, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
