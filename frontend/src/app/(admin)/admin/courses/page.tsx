"use client";

import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable, type Column } from "@/components/shared/data-table";
import { coursesService } from "@/services/courses.service";
import { formatNumber, formatPrice } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";
import type { Course } from "@/types";

export default function AdminCoursesPage() {
  const t = useT();
  const { data, isLoading } = useQuery({
    queryKey: ["admin", "courses"],
    queryFn: () => coursesService.list({ pageSize: 50 }),
  });

  const columns: Column<Course>[] = [
    {
      key: "title",
      header: t("admin.colCourse"),
      render: (c) => (
        <div className="flex items-center gap-3">
          <div className={`h-9 w-12 rounded bg-gradient-to-br ${c.thumbnailColor}`} />
          <div>
            <div className="font-semibold">{c.title}</div>
            <div className="text-xs text-muted-foreground">{t(`cat.${c.category.replace(/\s/g, "")}`)}</div>
          </div>
        </div>
      ),
    },
    { key: "instructor", header: t("admin.colInstructor"), accessor: (c) => c.instructor.name },
    { key: "students", header: t("admin.colStudents"), accessor: (c) => formatNumber(c.studentCount) },
    { key: "price", header: t("admin.colPrice"), accessor: (c) => formatPrice(c.price) },
    { key: "status", header: t("admin.colStatus"), render: () => <Badge variant="success">{t("admin.published")}</Badge> },
    {
      key: "actions",
      header: t("admin.colActions"),
      align: "right",
      render: () => (
        <Button variant="ghost" size="sm" className="text-rose-600">
          {t("admin.unpublish")}
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">{t("admin.courseManagement")}</h1>
      <DataTable columns={columns} data={data?.items ?? []} isLoading={isLoading} rowKey={(c) => c.id} />
    </div>
  );
}
