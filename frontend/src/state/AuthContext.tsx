import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { authApi } from "../api/resources";
import { authToken } from "../api/client";
import type { User } from "../types/domain";

interface AuthValue {
  user: User | null;
  loading: boolean;
  login(username: string, password: string): Promise<void>;
  logout(): Promise<void>;
}
const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(Boolean(authToken.get()));
  useEffect(() => {
    const expired = () => {
      authToken.set("");
      setUser(null);
    };
    window.addEventListener("session-expired", expired);
    if (authToken.get())
      authApi
        .me()
        .then(setUser)
        .catch(expired)
        .finally(() => setLoading(false));
    else setLoading(false);
    return () => window.removeEventListener("session-expired", expired);
  }, []);
  const value = useMemo<AuthValue>(
    () => ({
      user,
      loading,
      login: async (username, password) => {
        const result = await authApi.login(username, password);
        authToken.set(result.access_token);
        setUser(result.user);
      },
      logout: async () => {
        try {
          await authApi.logout();
        } finally {
          authToken.set("");
          setUser(null);
        }
      },
    }),
    [user, loading],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used inside AuthProvider");
  return value;
}
