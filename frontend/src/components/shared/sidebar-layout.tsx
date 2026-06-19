"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { LucideIcon } from "lucide-react";
import { Logo } from "./logo";
import { cn } from "@/lib/utils";

export interface SidebarItem {
  href: string;
  label: string;
  icon: LucideIcon;
  badge?: string | number;
}

interface SidebarLayoutProps {
  items: SidebarItem[];
  title?: string;
  variant?: "light" | "dark";
  children: React.ReactNode;
  topbar?: React.ReactNode;
}

export function SidebarLayout({ items, title, variant = "light", children, topbar }: SidebarLayoutProps) {
  const pathname = usePathname();
  const dark = variant === "dark";

  return (
    <div className="flex min-h-screen">
      <aside
        className={cn(
          "hidden w-64 shrink-0 flex-col p-4 lg:flex",
          dark ? "bg-slate-900 text-slate-300" : "border-r bg-card",
        )}
      >
        <div className="flex items-center gap-2 px-2 py-3">
          <Logo light={dark} />
          {title && <span className={cn("text-xs font-semibold", dark ? "text-primary" : "text-primary")}>{title}</span>}
        </div>
        <nav className="mt-6 space-y-1 text-sm font-medium">
          {items.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2.5 transition-colors",
                  active
                    ? dark
                      ? "bg-white/10 text-white"
                      : "bg-accent text-accent-foreground"
                    : dark
                      ? "hover:bg-white/5"
                      : "text-muted-foreground hover:bg-secondary",
                )}
              >
                <item.icon className="size-4" />
                <span className="flex-1">{item.label}</span>
                {item.badge != null && (
                  <span className="rounded-full bg-rose-500 px-1.5 text-xs text-white">{item.badge}</span>
                )}
              </Link>
            );
          })}
        </nav>
      </aside>

      <div className="flex flex-1 flex-col">
        {topbar && <header className="flex h-16 items-center justify-between border-b bg-card px-6">{topbar}</header>}
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
