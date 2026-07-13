"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Logo } from "./logo";
import { ThemeToggle } from "./theme-toggle";
import { LanguageSwitcher } from "./language-switcher";
import { SkipLink } from "./skip-link";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

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

function SidebarNav({
  items,
  title,
  dark,
  onNavigate,
}: {
  items: SidebarItem[];
  title?: string;
  dark: boolean;
  onNavigate?: () => void;
}) {
  const pathname = usePathname();
  const t = useT();

  return (
    <>
      <div className="flex items-center gap-2 px-2 py-3">
        <Logo light={dark} />
        {title && <span className="text-xs font-semibold text-primary">{title}</span>}
      </div>
      <nav className="mt-6 space-y-1 text-sm font-medium">
        {items.map((item) => {
          const active = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              aria-current={active ? "page" : undefined}
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
      <div className="mt-auto flex items-center justify-between px-3 pt-4">
        <span className={cn("text-xs font-medium", dark ? "text-slate-400" : "text-muted-foreground")}>
          {t("lang.theme")}
        </span>
        <div className="flex items-center">
          <LanguageSwitcher />
          <ThemeToggle />
        </div>
      </div>
    </>
  );
}

export function SidebarLayout({ items, title, variant = "light", children, topbar }: SidebarLayoutProps) {
  const pathname = usePathname();
  const t = useT();
  const dark = variant === "dark";
  const [mobileOpen, setMobileOpen] = useState(false);

  // Close the mobile drawer whenever the route changes.
  useEffect(() => setMobileOpen(false), [pathname]);

  return (
    <div className="flex min-h-screen">
      <SkipLink />
      {/* Desktop sidebar */}
      <aside
        className={cn(
          "hidden w-64 shrink-0 flex-col p-4 lg:flex",
          dark ? "bg-slate-900 text-slate-300" : "border-r bg-card",
        )}
      >
        <SidebarNav items={items} title={title} dark={dark} />
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        {/* Mobile top bar: hamburger + drawer. Shown below lg where the sidebar is hidden. */}
        <header className="flex h-16 items-center gap-3 border-b bg-card px-4 lg:hidden">
          <Dialog open={mobileOpen} onOpenChange={setMobileOpen}>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" aria-label={t("nav.openMenu")}>
                <Menu className="size-5" />
              </Button>
            </DialogTrigger>
            <DialogContent
              className={cn(
                "left-0 top-0 flex h-full max-w-xs translate-x-0 translate-y-0 flex-col gap-0 rounded-none rounded-r-2xl p-4",
                dark ? "bg-slate-900 text-slate-300" : "bg-card",
              )}
            >
              <DialogTitle className="sr-only">{title ?? t("nav.menu")}</DialogTitle>
              <SidebarNav items={items} title={title} dark={dark} onNavigate={() => setMobileOpen(false)} />
            </DialogContent>
          </Dialog>
          <Logo />
        </header>

        {topbar && (
          <header className="hidden h-16 items-center justify-between border-b bg-card px-6 lg:flex">
            {topbar}
          </header>
        )}
        <main id="main-content" className="min-w-0 flex-1 p-4 sm:p-6">
          <div className="mx-auto w-full max-w-7xl">{children}</div>
        </main>
      </div>
    </div>
  );
}
