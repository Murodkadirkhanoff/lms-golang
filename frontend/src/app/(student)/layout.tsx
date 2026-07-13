"use client";

import Link from "next/link";
import { Award, Bell, GraduationCap, Heart, LayoutDashboard, Receipt, Settings, User } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";
import { useAuth, initials } from "@/providers/auth-provider";

export default function StudentLayout({ children }: { children: React.ReactNode }) {
  const t = useT();
  const { user } = useAuth();
  const items: SidebarItem[] = [
    { href: ROUTES.dashboard, label: t("nav.dashboard"), icon: LayoutDashboard },
    { href: ROUTES.myCourses, label: t("nav.myCourses"), icon: GraduationCap },
    { href: ROUTES.wishlist, label: t("nav.wishlist"), icon: Heart },
    { href: ROUTES.certificates, label: t("nav.certificates"), icon: Award },
    { href: ROUTES.purchases, label: t("nav.purchases"), icon: Receipt },
    { href: ROUTES.notifications, label: t("notif.title"), icon: Bell },
    { href: ROUTES.profile, label: t("nav.profile"), icon: User },
    { href: ROUTES.settings, label: t("nav.settings"), icon: Settings },
  ];

  return (
    <SidebarLayout
      items={items}
      topbar={
        <>
          <Button asChild variant="ghost" size="sm">
            <Link href={ROUTES.courses}>
              <GraduationCap className="size-4" /> {t("student.browseCourses")}
            </Link>
          </Button>
          <div className="flex items-center gap-2">
            <Avatar className="size-9 bg-indigo-200">
              <AvatarFallback className="bg-indigo-200 text-indigo-700">
                {user ? initials(user.name) : "?"}
              </AvatarFallback>
            </Avatar>
            <div className="hidden text-sm sm:block">
              <div className="font-semibold">{user?.name ?? "Guest"}</div>
              <div className="text-xs text-muted-foreground">{t("student.learner")}</div>
            </div>
          </div>
        </>
      }
    >
      {children}
    </SidebarLayout>
  );
}
