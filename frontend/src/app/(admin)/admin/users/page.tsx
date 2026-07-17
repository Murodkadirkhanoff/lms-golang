"use client";

import { useState } from "react";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Pagination } from "@/components/shared/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { adminService, type AdminUser } from "@/services/admin.service";
import { useAuth } from "@/providers/auth-provider";
import { useToast } from "@/providers/toast-provider";
import { formatDate } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

const ROLES: AdminUser["role"][] = ["student", "instructor", "admin"];

export default function AdminUsersPage() {
  const t = useT();
  const toast = useToast();
  const { user: me } = useAuth();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading } = useQuery({
    queryKey: ["admin", "users", page],
    queryFn: () => adminService.getUsers(page, pageSize),
    placeholderData: keepPreviousData,
  });

  const roleMutation = useMutation({
    mutationFn: ({ id, role }: { id: number; role: AdminUser["role"] }) =>
      adminService.updateRole(id, role),
    onSuccess: () => {
      toast.success(t("admin.roleUpdated"));
      queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : t("common.somethingWrong")),
  });

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
      key: "role",
      header: t("admin.colRole"),
      render: (u) =>
        // Admin o'z rolini o'zgartira olmaydi (backend ham rad etadi).
        u.id === me?.id ? (
          <Badge variant="secondary">{u.role}</Badge>
        ) : (
          <Select
            value={u.role}
            onValueChange={(role) => roleMutation.mutate({ id: u.id, role: role as AdminUser["role"] })}
          >
            <SelectTrigger className="h-8 w-32 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ROLES.map((r) => (
                <SelectItem key={r} value={r}>
                  {r}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
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
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">{t("admin.userManagement")}</h1>
      <DataTable columns={columns} data={data?.items ?? []} isLoading={isLoading} rowKey={(u) => u.id} />
      {data && (
        <Pagination page={data.page} pageSize={data.pageSize} total={data.total} onPageChange={setPage} />
      )}
    </div>
  );
}
