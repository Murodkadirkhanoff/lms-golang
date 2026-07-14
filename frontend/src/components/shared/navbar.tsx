"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Award,
  BookOpen,
  GraduationCap,
  Heart,
  LayoutDashboard,
  LayoutGrid,
  LogOut,
  Menu,
  Receipt,
  Search,
  Settings,
  ShoppingCart,
  Shield,
  User,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Logo } from "./logo";
import { ThemeToggle } from "./theme-toggle";
import { LanguageSwitcher } from "./language-switcher";
import { ROUTES } from "@/constants";
import { cn } from "@/lib/utils";
import { useCart } from "@/features/cart/cart-context";
import { useWishlist } from "@/features/wishlist/wishlist-context";
import { useAuth, initials } from "@/providers/auth-provider";
import { useToast } from "@/providers/toast-provider";
import { useT } from "@/providers/locale-provider";

// Guests only see the catalog; signed-in users additionally get
// "My Learning" and "Teach" (no role gating between the two).
const publicNavLinks = [
  { href: ROUTES.courses, key: "nav.browse", icon: BookOpen },
  { href: ROUTES.categories, key: "nav.categories", icon: LayoutGrid },
];

const authedNavLinks = [
  { href: ROUTES.dashboard, key: "nav.myLearning", icon: LayoutDashboard },
  { href: ROUTES.studio, key: "nav.teach", icon: GraduationCap },
];

const accountLinks = [
  { href: ROUTES.dashboard, key: "nav.dashboard", icon: LayoutDashboard },
  { href: ROUTES.myCourses, key: "nav.myCourses", icon: GraduationCap },
  { href: ROUTES.certificates, key: "nav.certificates", icon: Award },
  { href: ROUTES.purchases, key: "nav.purchases", icon: Receipt },
  { href: ROUTES.profile, key: "nav.profile", icon: User },
  { href: ROUTES.settings, key: "nav.settings", icon: Settings },
];

// Admin panel — only for users whose backend role is "admin".
const adminLink = { href: ROUTES.admin, key: "nav.admin", icon: Shield };

export function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const cart = useCart();
  const wishlist = useWishlist();
  const { user, logout, isLoading } = useAuth();
  const toast = useToast();
  const t = useT();
  const [search, setSearch] = useState("");
  const [mobileOpen, setMobileOpen] = useState(false);

  const navLinks = user ? [...publicNavLinks, ...authedNavLinks] : publicNavLinks;
  const menuLinks = user?.role === "admin" ? [...accountLinks, adminLink] : accountLinks;

  const submitSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setMobileOpen(false);
    router.push(`${ROUTES.courses}?search=${encodeURIComponent(search.trim())}`);
  };

  const handleLogout = () => {
    logout();
    toast.success(t("toast.loggedOut"));
    router.push(ROUTES.login);
  };

  const isActive = (href: string) => pathname === href || pathname.startsWith(href + "/");

  return (
    <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
      <nav className="mx-auto flex h-16 max-w-7xl items-center gap-4 px-4 sm:px-6">
        {/* Mobile menu trigger + drawer */}
        <Dialog open={mobileOpen} onOpenChange={setMobileOpen}>
          <DialogTrigger asChild>
            <Button variant="ghost" size="icon" className="lg:hidden" aria-label={t("nav.openMenu")}>
              <Menu className="size-5" />
            </Button>
          </DialogTrigger>
          <DialogContent className="left-0 top-0 h-full max-w-xs translate-x-0 translate-y-0 gap-0 rounded-none rounded-r-2xl p-0">
            <DialogTitle className="sr-only">{t("nav.menu")}</DialogTitle>
            <div className="flex h-full flex-col p-4">
              <div className="px-2 py-2">
                <Logo />
              </div>
              <form onSubmit={submitSearch} className="relative mt-4">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder={t("nav.searchPlaceholder")}
                  className="h-10 w-full rounded-full border border-input bg-secondary/60 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </form>
              <div className="mt-4 space-y-1">
                {navLinks.map((link) => (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={() => setMobileOpen(false)}
                    aria-current={isActive(link.href) ? "page" : undefined}
                    className={cn(
                      "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                      isActive(link.href)
                        ? "bg-accent text-accent-foreground"
                        : "text-muted-foreground hover:bg-secondary hover:text-foreground",
                    )}
                  >
                    <link.icon className="size-4" />
                    {t(link.key)}
                  </Link>
                ))}
              </div>
              <div className="mt-auto flex items-center justify-between border-t px-1 pt-4">
                <LanguageSwitcher />
                <ThemeToggle />
              </div>
            </div>
          </DialogContent>
        </Dialog>

        <div className="flex items-center gap-8">
          <Logo />
          <div className="hidden items-center gap-1 lg:flex">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                aria-current={isActive(link.href) ? "page" : undefined}
                className={cn(
                  "rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  isActive(link.href)
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {t(link.key)}
              </Link>
            ))}
          </div>
        </div>

        {/* Search */}
        <form onSubmit={submitSearch} className="relative ml-auto hidden max-w-md flex-1 md:block">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("nav.searchPlaceholder")}
            className="h-10 w-full rounded-full border border-input bg-secondary/60 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </form>

        <div className="ml-auto flex items-center gap-1 md:ml-0">
          <div className="hidden sm:flex">
            <LanguageSwitcher />
            <ThemeToggle />
          </div>
          <IconLink href={ROUTES.wishlist} label={t("nav.wishlist")} count={wishlist.count}>
            <Heart className="size-5" />
          </IconLink>
          <IconLink href={ROUTES.cart} label={t("nav.cart")} count={cart.count}>
            <ShoppingCart className="size-5" />
          </IconLink>

          {user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="ml-1 rounded-full focus:outline-none focus:ring-2 focus:ring-ring" aria-label={t("nav.account")}>
                  <Avatar className="bg-indigo-200">
                    <AvatarFallback className="bg-indigo-200 text-indigo-700">
                      {initials(user.name)}
                    </AvatarFallback>
                  </Avatar>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>
                  <div className="font-semibold">{user.name}</div>
                  <div className="text-xs text-muted-foreground">{user.email}</div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                {menuLinks.map((link) => (
                  <DropdownMenuItem key={link.href} asChild>
                    <Link href={link.href}>
                      <link.icon className="size-4 text-muted-foreground" />
                      {t(link.key)}
                    </Link>
                  </DropdownMenuItem>
                ))}
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="size-4 text-muted-foreground" />
                  {t("common.logout")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            !isLoading && (
              <div className="ml-1 flex items-center gap-2">
                <Button asChild variant="ghost" size="sm">
                  <Link href={ROUTES.login}>{t("common.login")}</Link>
                </Button>
                <Button asChild size="sm">
                  <Link href={ROUTES.register}>{t("auth.signUp")}</Link>
                </Button>
              </div>
            )
          )}
        </div>
      </nav>
    </header>
  );
}

function IconLink({
  href,
  label,
  count,
  children,
}: {
  href: string;
  label: string;
  count: number;
  children: React.ReactNode;
}) {
  return (
    <Button asChild variant="ghost" size="icon" className="relative">
      <Link href={href} aria-label={label}>
        {children}
        {count > 0 && (
          <span className="absolute -right-0.5 -top-0.5 grid size-4 place-items-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
            {count}
          </span>
        )}
      </Link>
    </Button>
  );
}
