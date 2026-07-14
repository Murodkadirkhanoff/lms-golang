"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { ROUTES } from "@/constants";
import { useAuth } from "@/providers/auth-provider";

/**
 * Guards a protected subtree: while the stored session is loading nothing is
 * rendered (avoids a signed-out flash), and unauthenticated visitors are
 * redirected to the login page with the original destination preserved.
 */
export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace(`${ROUTES.login}?next=${encodeURIComponent(pathname)}`);
    }
  }, [isLoading, isAuthenticated, router, pathname]);

  if (isLoading || !isAuthenticated) return null;

  return <>{children}</>;
}
