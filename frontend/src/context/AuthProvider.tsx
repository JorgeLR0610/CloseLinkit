import { useEffect, useState, useCallback, type ReactNode } from "react";
import type { User } from "../types/auth";
import * as authService from "../services/auth";
import { claimStoredURLs } from "../services/urls";
import { AuthContext } from "./AuthContext";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Restore session on initial application mount using HttpOnly cookie
  useEffect(() => {
    let isMounted = true;

    async function restoreSession() {
      try {
        const res = await authService.refreshToken();
        if (isMounted) {
          try {
            await claimStoredURLs(res.access_token);
          } catch {
            // ignore
          }
          setAccessToken(res.access_token);
          setUser(res.user);
        }
      } catch {
        if (isMounted) {
          setAccessToken(null);
          setUser(null);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    void restoreSession();

    return () => {
      isMounted = false;
    };
  }, []);

  const login = useCallback(async (email: string, password: string): Promise<void> => {
    const res = await authService.login({ email, password });
    try {
      await claimStoredURLs(res.access_token);
    } catch {
      // ignore
    }
    setAccessToken(res.access_token);
    setUser(res.user);
  }, []);

  const register = useCallback(
    async (email: string, password: string): Promise<void> => {
      await authService.register({ email, password });
      // Auto-login with same credentials
      await login(email, password);
    },
    [login],
  );

  const logout = useCallback(async (): Promise<void> => {
    try {
      await authService.logout();
    } finally {
      setAccessToken(null);
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        accessToken,
        isLoading,
        isAuthenticated: !!user && !!accessToken,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
