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

const schema = z.object({
  name_en: z.string().min(1, "Required").max(100),
  name_uz: z.string().min(1, "Required").max(100),
  name_ru: z.string().min(1, "Required").max(100),
});
type Values = z.infer<typeof schema>;

export default function AdminCategoriesPage() {
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
    mutationFn: (values: Values) => categoriesService.create(values),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["categories"] });
      reset();
      setOpen(false);
    },
    onError: (err) => setServerError(err instanceof Error ? err.message : "Failed to create category"),
  });

  const nameEn = watch("name_en") ?? "";

  const columns: Column<Category>[] = [
    { key: "name", header: "Name (EN)", render: (c) => <span className="font-semibold">{c.nameEn}</span> },
    { key: "uz", header: "Uzbek", accessor: (c) => c.nameUz },
    { key: "ru", header: "Russian", accessor: (c) => c.nameRu },
    { key: "slug", header: "Slug", render: (c) => <Badge variant="secondary">{c.slug}</Badge> },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-extrabold">Category management</h1>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="size-4" /> New category
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create category</DialogTitle>
              <DialogDescription>The slug is generated from the English name.</DialogDescription>
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
              <FormField label="Name (English)" placeholder="Development" error={errors.name_en?.message} {...register("name_en")} />
              <FormField label="Name (Uzbek)" placeholder="Dasturlash" error={errors.name_uz?.message} {...register("name_uz")} />
              <FormField label="Name (Russian)" placeholder="Разработка" error={errors.name_ru?.message} {...register("name_ru")} />
              <p className="text-xs text-muted-foreground">
                Slug preview: <span className="font-mono">{slugify(nameEn) || "—"}</span>
              </p>
            </form>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" form="category-form" disabled={mutation.isPending}>
                {mutation.isPending ? "Creating…" : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} rowKey={(c) => c.id} />
    </div>
  );
}
