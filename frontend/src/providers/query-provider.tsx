"use client";

import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "@/providers/theme-provider";
import { LocaleProvider } from "@/providers/locale-provider";
import { AuthProvider } from "@/providers/auth-provider";
import { ToastProvider } from "@/providers/toast-provider";
import { CartProvider } from "@/features/cart/cart-context";
import { WishlistProvider } from "@/features/wishlist/wishlist-context";

export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      }),
  );

  return (
    <ThemeProvider>
      <LocaleProvider>
        <ToastProvider>
          <QueryClientProvider client={client}>
            <AuthProvider>
              <WishlistProvider>
                <CartProvider>{children}</CartProvider>
              </WishlistProvider>
            </AuthProvider>
          </QueryClientProvider>
        </ToastProvider>
      </LocaleProvider>
    </ThemeProvider>
  );
}
