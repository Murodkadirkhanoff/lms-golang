"use client";

import { BookOpen, FolderTree, LayoutDashboard, Users } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { ROUTES } from "@/constants";

const items: SidebarItem[] = [
  { href: ROUTES.admin, label: "Overview", icon: LayoutDashboard },
  { href: ROUTES.adminUsers, label: "Users", icon: Users },
  { href: ROUTES.adminCourses, label: "Courses", icon: BookOpen },
  { href: ROUTES.adminCategories, label: "Categories", icon: FolderTree },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarLayout
      items={items}
      title="Admin"
      variant="dark"
      topbar={<h1 className="text-lg font-bold">Platform administration</h1>}
    >
      {children}
    </SidebarLayout>
  );
}
