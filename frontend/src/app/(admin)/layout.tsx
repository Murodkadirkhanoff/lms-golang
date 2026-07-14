"use client";

import { BookOpen, FolderTree, LayoutDashboard, Users } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { RequireAuth } from "@/components/shared/require-auth";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const t = useT();
  const items: SidebarItem[] = [
    { href: ROUTES.admin, label: t("admin.overview"), icon: LayoutDashboard },
    { href: ROUTES.adminUsers, label: t("admin.users"), icon: Users },
    { href: ROUTES.adminCourses, label: t("admin.courses"), icon: BookOpen },
    { href: ROUTES.adminCategories, label: t("admin.categories"), icon: FolderTree },
  ];

  return (
    <RequireAuth>
      <SidebarLayout
        items={items}
        title={t("admin.title")}
        variant="dark"
        topbar={<h1 className="text-lg font-bold">{t("admin.administration")}</h1>}
      >
        {children}
      </SidebarLayout>
    </RequireAuth>
  );
}
