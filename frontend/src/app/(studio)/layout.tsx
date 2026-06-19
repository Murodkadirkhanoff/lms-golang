"use client";

import Link from "next/link";
import { BarChart3, BookOpen, LayoutDashboard, Plus } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/constants";

const items: SidebarItem[] = [
  { href: ROUTES.studio, label: "Dashboard", icon: LayoutDashboard },
  { href: ROUTES.studioCourses, label: "My Courses", icon: BookOpen },
  { href: ROUTES.studioAnalytics, label: "Analytics", icon: BarChart3 },
];

export default function StudioLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarLayout
      items={items}
      title="Studio"
      variant="dark"
      topbar={
        <>
          <Button asChild variant="ghost" size="sm">
            <Link href={ROUTES.dashboard}>← Back to learning</Link>
          </Button>
          <Button asChild>
            <Link href={ROUTES.studioCourseNew}>
              <Plus className="size-4" /> New course
            </Link>
          </Button>
        </>
      }
    >
      {children}
    </SidebarLayout>
  );
}
