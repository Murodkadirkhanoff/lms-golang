"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { authService, getStoredUser } from "@/services/auth.service";
import type { User } from "@/types";

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  /** True until the client has read the stored session (avoids SSR mismatch). */
  isLoading: boolean;
  setUser: (user: User | null) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  // Server and first client render start signed-out so markup matches;
  // the stored session (if any) is reconciled on mount.
  const [user, setUserState] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    setUserState(getStoredUser());
    setIsLoading(false);
  }, []);

  const setUser = useCallback((next: User | null) => setUserState(next), []);

  const logout = useCallback(() => {
    authService.logout();
    setUserState(null);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ user, isAuthenticated: user !== null, isLoading, setUser, logout }),
    [user, isLoading, setUser, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

/** Initials for avatar fallbacks, e.g. "Amir Karimov" → "AK". */
export function initials(name: string): string {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? "")
    .join("");
}
