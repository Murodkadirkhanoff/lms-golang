"use client";

import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable, type Column } from "@/components/shared/data-table";
import { adminService, type AdminUser } from "@/services/admin.service";
import { formatDate } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

export default function AdminUsersPage() {
  const t = useT();
  const { data, isLoading } = useQuery({ queryKey: ["admin", "users"], queryFn: adminService.getUsers });

  const columns: Column<AdminUser>[] = [
    {
      key: "user",
      header: t("admin.colUser"),
      render: (u) => (
        <div className="flex items-center gap-3">
          <div className={`size-8 rounded-full ${u.avatarColor}`} />
          <div>
            <div className="font-semibold">{u.name}</div>
            <div className="text-xs text-muted-foreground">{u.email}</div>
          </div>
        </div>
      ),
    },
    {
      key: "activity",
      header: t("admin.colActivity"),
      render: (u) => (
        <span className="text-xs text-muted-foreground">
          {t("admin.activityMeta", { created: u.coursesCreated, enrolled: u.coursesEnrolled })}
        </span>
      ),
    },
    { key: "joined", header: t("admin.colJoined"), accessor: (u) => formatDate(u.joinedAt) },
    {
      key: "status",
      header: t("admin.colStatus"),
      render: (u) =>
        u.status === "active" ? <Badge variant="success">{t("admin.active")}</Badge> : <Badge variant="destructive">{t("admin.suspended")}</Badge>,
    },
    {
      key: "actions",
      header: t("admin.colActions"),
      align: "right",
      render: (u) => (
        <Button variant="ghost" size="sm" className={u.status === "active" ? "text-rose-600" : "text-emerald-600"}>
          {u.status === "active" ? t("admin.suspend") : t("admin.restore")}
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">{t("admin.userManagement")}</h1>
      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} rowKey={(u) => u.id} />
    </div>
  );
}
