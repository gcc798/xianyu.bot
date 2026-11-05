import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { PropsWithChildren } from "react";
import { getCurrentUser, login, logout, readAuthSession, storeAuthSession } from "../api/auth";
import type { AuthSession, LoginRequest } from "../models/auth";

type AuthContextValue = {
  session: AuthSession | null;
  isReady: boolean;
  signIn: (input: LoginRequest) => Promise<void>;
  signOut: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession | null>(() => readAuthSession());
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    if (!session) {
      setIsReady(true);
      return;
    }
    let active = true;
    void getCurrentUser()
      .then(({ userInfo }) => {
        if (!active) return;
        const renewed = readAuthSession();
        const current = { ...(renewed || session), user: userInfo };
        storeAuthSession(current);
        setSession(current);
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setIsReady(true);
      });
    return () => { active = false; };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isReady,
      signIn: async (input) => setSession(await login(input)),
      signOut: async () => {
        await logout();
        setSession(null);
      },
    }),
    [isReady, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth 必须在 AuthProvider 内使用");
  return value;
}
