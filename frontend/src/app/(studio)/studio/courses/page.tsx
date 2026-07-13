"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DataTable, type Column } from "@/components/shared/data-table";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { coursesService } from "@/services/courses.service";
import { ROUTES } from "@/constants";
import { formatNumber, formatPrice } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";
import type { Course } from "@/types";

export default function StudioCoursesPage() {
  const t = useT();
  const { data, isLoading } = useQuery({
    queryKey: ["studio", "courses"],
    queryFn: () => coursesService.list({ pageSize: 50 }),
  });
  const [toDelete, setToDelete] = useState<Course | null>(null);

  const columns: Column<Course>[] = [
    {
      key: "title",
      header: t("studio.colCourse"),
      render: (c) => (
        <div className="flex items-center gap-3">
          <div className={`h-9 w-12 rounded bg-gradient-to-br ${c.thumbnailColor}`} />
          <span className="font-semibold">{c.title}</span>
        </div>
      ),
    },
    {
      key: "status",
      header: t("studio.colStatus"),
      render: (c) =>
        c.isPublished ? <Badge variant="success">{t("studio.published")}</Badge> : <Badge variant="warning">{t("studio.draft")}</Badge>,
    },
    { key: "students", header: t("studio.colStudents"), accessor: (c) => formatNumber(c.studentCount) },
    { key: "rating", header: t("studio.colRating"), accessor: (c) => `${c.rating} ★` },
    { key: "revenue", header: t("studio.colRevenue"), render: (c) => <span className="font-semibold">{formatPrice(c.price * c.studentCount)}</span> },
    {
      key: "actions",
      header: t("studio.colActions"),
      align: "right",
      render: (c) => (
        <div className="flex justify-end gap-2">
          <Button asChild variant="ghost" size="sm">
            <Link href={ROUTES.studioCourseEdit(c.id)}>{t("studio.edit")}</Link>
          </Button>
          <Button variant="ghost" size="sm" className="text-rose-600" onClick={() => setToDelete(c)}>
            {t("studio.delete")}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold">{t("studio.myCourses")}</h1>
        <Button asChild>
          <Link href={ROUTES.studioCourseNew}>
            <Plus className="size-4" /> {t("studio.newCourse")}
          </Link>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={data?.items ?? []}
        isLoading={isLoading}
        rowKey={(c) => c.id}
        emptyMessage={t("studio.noCourses")}
      />

      {/* Delete confirmation modal */}
      <Dialog open={!!toDelete} onOpenChange={(open) => !open && setToDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("studio.deleteTitle")}</DialogTitle>
            <DialogDescription>
              {t("studio.deleteDesc", { title: toDelete?.title ?? "" })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">{t("common.cancel")}</Button>
            </DialogClose>
            <Button variant="destructive" onClick={() => setToDelete(null)}>
              {t("studio.deleteCourse")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
