"use client";

import Link from "next/link";
import { BarChart3, BookOpen, LayoutDashboard, Plus } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { RequireAuth } from "@/components/shared/require-auth";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function StudioLayout({ children }: { children: React.ReactNode }) {
  const t = useT();
  const items: SidebarItem[] = [
    { href: ROUTES.studio, label: t("studio.dashboard"), icon: LayoutDashboard },
    { href: ROUTES.studioCourses, label: t("studio.myCourses"), icon: BookOpen },
    { href: ROUTES.studioAnalytics, label: t("studio.analytics"), icon: BarChart3 },
  ];

  return (
    <RequireAuth>
    <SidebarLayout
      items={items}
      title={t("studio.title")}
      variant="dark"
      topbar={
        <>
          <Button asChild variant="ghost" size="sm">
            <Link href={ROUTES.dashboard}>{t("studio.backToLearning")}</Link>
          </Button>
          <Button asChild>
            <Link href={ROUTES.studioCourseNew}>
              <Plus className="size-4" /> {t("studio.newCourse")}
            </Link>
          </Button>
        </>
      }
    >
      {children}
    </SidebarLayout>
    </RequireAuth>
  );
}
