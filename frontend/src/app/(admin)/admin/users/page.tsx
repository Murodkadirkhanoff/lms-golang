"use client";

import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable, type Column } from "@/components/shared/data-table";
import { adminService, type AdminUser } from "@/services/admin.service";
import { formatDate } from "@/lib/utils";

export default function AdminUsersPage() {
  const { data, isLoading } = useQuery({ queryKey: ["admin", "users"], queryFn: adminService.getUsers });

  const columns: Column<AdminUser>[] = [
    {
      key: "user",
      header: "User",
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
      header: "Activity",
      render: (u) => (
        <span className="text-xs text-muted-foreground">
          {u.coursesCreated} created · {u.coursesEnrolled} enrolled
        </span>
      ),
    },
    { key: "joined", header: "Joined", accessor: (u) => formatDate(u.joinedAt) },
    {
      key: "status",
      header: "Status",
      render: (u) =>
        u.status === "active" ? <Badge variant="success">Active</Badge> : <Badge variant="destructive">Suspended</Badge>,
    },
    {
      key: "actions",
      header: "Actions",
      align: "right",
      render: (u) => (
        <Button variant="ghost" size="sm" className={u.status === "active" ? "text-rose-600" : "text-emerald-600"}>
          {u.status === "active" ? "Suspend" : "Restore"}
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">User management</h1>
      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} rowKey={(u) => u.id} />
    </div>
  );
}
