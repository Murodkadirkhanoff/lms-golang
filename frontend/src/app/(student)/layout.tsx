"use client";

import Link from "next/link";
import { Award, LayoutDashboard, GraduationCap, User } from "lucide-react";
import { SidebarLayout, type SidebarItem } from "@/components/shared/sidebar-layout";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/constants";

const items: SidebarItem[] = [
  { href: ROUTES.dashboard, label: "Dashboard", icon: LayoutDashboard },
  { href: ROUTES.certificates, label: "Certificates", icon: Award },
  { href: ROUTES.profile, label: "Profile", icon: User },
];

export default function StudentLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarLayout
      items={items}
      topbar={
        <>
          <Button asChild variant="ghost" size="sm">
            <Link href={ROUTES.courses}>
              <GraduationCap className="size-4" /> Browse courses
            </Link>
          </Button>
          <div className="flex items-center gap-2">
            <div className="size-9 rounded-full bg-indigo-200" />
            <div className="hidden text-sm sm:block">
              <div className="font-semibold">Amir K.</div>
              <div className="text-xs text-muted-foreground">Learner</div>
            </div>
          </div>
        </>
      }
    >
      {children}
    </SidebarLayout>
  );
}
