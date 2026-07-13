"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { FormField } from "@/components/ui/form-field";
import { DataTable, type Column } from "@/components/shared/data-table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { categoriesService } from "@/services/categories.service";
import type { Category } from "@/types";
import { slugify } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

const schema = z.object({
  name_en: z.string().min(1, "Required").max(100),
  name_uz: z.string().min(1, "Required").max(100),
  name_ru: z.string().min(1, "Required").max(100),
  // Raw <select> value: "" means top-level, otherwise the parent id as a string.
  parent_id: z.string().optional(),
});
type Values = z.infer<typeof schema>;

export default function AdminCategoriesPage() {
  const t = useT();
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [serverError, setServerError] = useState("");

  const { data, isLoading } = useQuery({ queryKey: ["categories"], queryFn: categoriesService.list });

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors },
  } = useForm<Values>({ resolver: zodResolver(schema) });

  const mutation = useMutation({
    mutationFn: (values: Values) =>
      categoriesService.create({
        name_en: values.name_en,
        name_uz: values.name_uz,
        name_ru: values.name_ru,
        parent_id: values.parent_id ? Number(values.parent_id) : null,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["categories"] });
      reset();
      setOpen(false);
    },
    onError: (err) => setServerError(err instanceof Error ? err.message : "Failed to create category"),
  });

  const nameEn = watch("name_en") ?? "";

  const categories = data ?? [];
  const parents = categories.filter((c) => c.parentId == null);
  const nameById = new Map(categories.map((c) => [c.id, c.nameEn]));
  // Order rows so each parent is immediately followed by its children.
  const ordered = parents.flatMap((p) => [p, ...categories.filter((c) => c.parentId === p.id)]);

  const columns: Column<Category>[] = [
    {
      key: "name",
      header: t("admin.colNameEn"),
      render: (c) => (
        <span className={c.parentId == null ? "font-semibold" : "pl-4 text-muted-foreground"}>
          {c.parentId == null ? c.nameEn : `— ${c.nameEn}`}
        </span>
      ),
    },
    { key: "uz", header: t("admin.colUzbek"), accessor: (c) => c.nameUz },
    { key: "ru", header: t("admin.colRussian"), accessor: (c) => c.nameRu },
    {
      key: "parent",
      header: t("admin.colParent"),
      render: (c) =>
        c.parentId == null ? (
          <Badge>{t("admin.topLevel")}</Badge>
        ) : (
          <span className="text-sm text-muted-foreground">{nameById.get(c.parentId) ?? "—"}</span>
        ),
    },
    { key: "slug", header: t("admin.colSlug"), render: (c) => <Badge variant="secondary">{c.slug}</Badge> },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold">{t("admin.categoryManagement")}</h1>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="size-4" /> {t("admin.newCategory")}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("admin.createCategory")}</DialogTitle>
              <DialogDescription>{t("admin.createCategoryDesc")}</DialogDescription>
            </DialogHeader>
            <form
              id="category-form"
              onSubmit={handleSubmit((v) => {
                setServerError("");
                mutation.mutate(v);
              })}
              className="space-y-4"
            >
              {serverError && <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{serverError}</div>}
              <FormField label={t("admin.nameEn")} placeholder="Development" error={errors.name_en?.message} {...register("name_en")} />
              <FormField label={t("admin.nameUz")} placeholder="Dasturlash" error={errors.name_uz?.message} {...register("name_uz")} />
              <FormField label={t("admin.nameRu")} placeholder="Разработка" error={errors.name_ru?.message} {...register("name_ru")} />
              <div className="space-y-1.5">
                <label className="text-sm font-medium" htmlFor="parent_id">
                  {t("admin.parent")}
                </label>
                <select
                  id="parent_id"
                  defaultValue=""
                  className="flex h-10 w-full rounded-lg border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  {...register("parent_id")}
                >
                  <option value="">{t("admin.topLevel")}</option>
                  {parents.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.nameEn}
                    </option>
                  ))}
                </select>
              </div>
              <p className="text-xs text-muted-foreground">
                {t("admin.slugPreview")} <span className="font-mono">{slugify(nameEn) || "—"}</span>
              </p>
            </form>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {t("common.cancel")}
              </Button>
              <Button type="submit" form="category-form" disabled={mutation.isPending}>
                {mutation.isPending ? t("admin.creating") : t("admin.create")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <DataTable columns={columns} data={ordered} isLoading={isLoading} rowKey={(c) => c.id} />
    </div>
  );
}
