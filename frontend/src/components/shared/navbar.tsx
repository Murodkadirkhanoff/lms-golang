"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { GraduationCap, LayoutDashboard, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Logo } from "./logo";
import { ROUTES } from "@/constants";
import { cn } from "@/lib/utils";

// No role gating: every signed-in user sees both "My Learning" and "Teach".
const navLinks = [
  { href: ROUTES.courses, label: "Browse" },
  { href: ROUTES.dashboard, label: "My Learning", icon: LayoutDashboard },
  { href: ROUTES.studio, label: "Teach", icon: GraduationCap },
];

export function Navbar() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
      <nav className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <Logo />
          <div className="hidden items-center gap-1 md:flex">
            {navLinks.map((link) => {
              const active = pathname === link.href || pathname.startsWith(link.href + "/");
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={cn(
                    "rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                    active ? "bg-accent text-accent-foreground" : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button asChild variant="ghost" size="icon" className="hidden sm:inline-flex">
            <Link href={ROUTES.courses} aria-label="Search courses">
              <Search />
            </Link>
          </Button>
          <Button asChild variant="ghost" className="hidden sm:inline-flex">
            <Link href={ROUTES.login}>Log in</Link>
          </Button>
          <Button asChild>
            <Link href={ROUTES.register}>Get started</Link>
          </Button>
        </div>
      </nav>
    </header>
  );
}
