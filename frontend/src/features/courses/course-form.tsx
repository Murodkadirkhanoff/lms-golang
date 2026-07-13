"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Controller, FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ImagePlus, Info } from "lucide-react";
import { CurriculumBuilder } from "./curriculum-builder";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { coursesService } from "@/services/courses.service";
import { categoriesService } from "@/services/categories.service";
import { courseSchema, type CourseFormValues } from "./course-schema";
import { LANGUAGES, ROUTES } from "@/constants";
import { formatPrice, slugify } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

export function CourseForm() {
  const router = useRouter();
  const t = useT();
  const [serverError, setServerError] = useState("");

  const { data: categories } = useQuery({ queryKey: ["categories"], queryFn: categoriesService.list });

  const methods = useForm<CourseFormValues>({
    resolver: zodResolver(courseSchema),
    defaultValues: { lang: "uz", price: 0, isPublished: false, categoryId: null, description: "", modules: [] },
  });
  const {
    register,
    handleSubmit,
    control,
    watch,
    formState: { errors },
  } = methods;

  const title = watch("title") ?? "";
  const description = watch("description") ?? "";
  const price = watch("price");
  const slug = slugify(title);

  const mutation = useMutation({
    mutationFn: (values: CourseFormValues) =>
      coursesService.create({
        title: values.title,
        description: values.description ?? "",
        categoryId: values.categoryId,
        lang: values.lang,
        price: values.price,
        isPublished: values.isPublished,
        modules: values.modules.map((m) => ({
          title: m.title,
          lessons: m.lessons.map((l) => ({
            title: l.title,
            type: l.type,
            contentUrl: l.type === "video" ? l.contentUrl ?? "" : "",
            content: l.type === "text" ? l.content ?? "" : "",
            durationSeconds: Math.round((l.durationMinutes ?? 0) * 60),
            price: l.price,
            isFree: l.isFree,
          })),
        })),
      }),
    onSuccess: () => router.push(ROUTES.studioCourses),
    onError: (err) => setServerError(err instanceof Error ? err.message : "Failed to create course"),
  });

  const submit = (publish: boolean) =>
    handleSubmit((values) => {
      setServerError("");
      mutation.mutate({ ...values, isPublished: publish });
    });

  return (
    <FormProvider {...methods}>
    <form className="mx-auto grid w-full max-w-5xl gap-6 lg:grid-cols-3">
      <div className="min-w-0 space-y-6 lg:col-span-2">
        {serverError && <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{serverError}</div>}

        {/* Basic info */}
        <Card>
          <CardHeader>
            <CardTitle>{t("cf.basicInfo")}</CardTitle>
            <CardDescription>{t("cf.basicInfoDesc")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-1.5">
              <Label htmlFor="title">
                {t("cf.courseTitle")} <span className="text-rose-500">*</span>
              </Label>
              <Input id="title" placeholder={t("cf.titlePlaceholder")} {...register("title")} />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{errors.title?.message ?? t("cf.titleHint")}</span>
                <span>{title.length} / 200</span>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t("cf.urlSlug")}</Label>
              <div className="flex items-center overflow-hidden rounded-lg border bg-secondary/50">
                <span className="shrink-0 border-r bg-secondary px-3 py-2 text-sm text-muted-foreground max-sm:text-xs">
                  learnhub.com/courses/
                </span>
                <span className="min-w-0 truncate px-3 py-2 text-sm text-muted-foreground">{slug || "your-course-title"}</span>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">{t("cf.description")}</Label>
              <Textarea id="description" rows={5} placeholder={t("cf.descPlaceholder")} {...register("description")} />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{errors.description?.message ?? t("cf.descHint")}</span>
                <span>{description.length} / 5000</span>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label>{t("cf.category")}</Label>
                <Controller
                  control={control}
                  name="categoryId"
                  render={({ field }) => (
                    <Select
                      value={field.value ? String(field.value) : ""}
                      onValueChange={(v) => field.onChange(Number(v))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder={t("cf.selectCategory")} />
                      </SelectTrigger>
                      <SelectContent>
                        {(categories ?? [])
                          .filter((c) => c.parentId == null)
                          .map((parent) => {
                            const children = (categories ?? []).filter((c) => c.parentId === parent.id);
                            if (children.length === 0) return null;
                            return (
                              <SelectGroup key={parent.id}>
                                <SelectLabel>{parent.nameEn}</SelectLabel>
                                {children.map((child) => (
                                  <SelectItem key={child.id} value={String(child.id)}>
                                    {child.nameEn}
                                  </SelectItem>
                                ))}
                              </SelectGroup>
                            );
                          })}
                      </SelectContent>
                    </Select>
                  )}
                />
                {errors.categoryId && <p className="text-xs text-destructive">{errors.categoryId.message}</p>}
              </div>

              <div className="space-y-1.5">
                <Label>
                  {t("cf.language")} <span className="text-rose-500">*</span>
                </Label>
                <Controller
                  control={control}
                  name="lang"
                  render={({ field }) => (
                    <div className="flex gap-2">
                      {LANGUAGES.map((l) => (
                        <button
                          type="button"
                          key={l.value}
                          onClick={() => field.onChange(l.value)}
                          className={`flex-1 rounded-lg border py-2.5 text-sm font-medium ${
                            field.value === l.value ? "border-primary bg-accent text-accent-foreground" : ""
                          }`}
                        >
                          {l.value.toUpperCase()}
                        </button>
                      ))}
                    </div>
                  )}
                />
              </div>
            </div>

            <p className="flex items-center gap-2 rounded-lg bg-secondary/60 px-3 py-2 text-xs text-muted-foreground">
              <Info className="size-4 shrink-0" />
              {t("cf.instructorNote")}
            </p>
          </CardContent>
        </Card>

        {/* Curriculum: modules -> lessons */}
        <CurriculumBuilder />

        {/* Thumbnail */}
        <Card>
          <CardHeader>
            <CardTitle>{t("cf.thumbnail")}</CardTitle>
            <CardDescription>{t("cf.thumbnailDesc")}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid place-items-center rounded-xl border-2 border-dashed p-8 text-center">
              <ImagePlus className="size-8 text-muted-foreground" />
              <p className="mt-2 text-sm font-semibold">
                {t("cf.dragDrop")} <span className="text-primary">{t("cf.browse")}</span>
              </p>
              <p className="text-xs text-muted-foreground">{t("cf.fileHint")}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Sidebar: pricing + publish */}
      <div className="min-w-0 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{t("cf.pricing")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="price">{t("cf.priceUzs")}</Label>
              <div className="flex items-center overflow-hidden rounded-lg border">
                <input
                  id="price"
                  type="number"
                  step="1000"
                  min={0}
                  className="w-full min-w-0 flex-1 px-3 py-2 text-sm focus:outline-none"
                  {...register("price", { valueAsNumber: true })}
                />
                <span className="shrink-0 border-l bg-secondary px-3 py-2 text-muted-foreground">{t("common.sum")}</span>
              </div>
              <p className="text-xs text-muted-foreground">{t("cf.priceHint")}</p>
              {errors.price && <p className="text-xs text-destructive">{errors.price.message}</p>}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("cf.publish")}</CardTitle>
            <CardDescription>{t("cf.publishDesc")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button type="button" className="w-full" disabled={mutation.isPending} onClick={submit(true)}>
              {mutation.isPending ? t("cf.saving") : t("cf.publishCourse")}
            </Button>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={mutation.isPending}
              onClick={submit(false)}
            >
              {t("cf.saveDraft")}
            </Button>
          </CardContent>
        </Card>

        {/* Live preview */}
        <Card className="overflow-hidden">
          <div className="aspect-video bg-gradient-to-br from-indigo-400 to-violet-500" />
          <CardContent className="p-4">
            <h3 className="line-clamp-2 text-sm font-bold">{title || t("cf.yourTitle")}</h3>
            <p className="mt-0.5 text-xs text-muted-foreground">{t("cf.byYou")}</p>
            <div className="mt-2 font-extrabold">{Number(price) > 0 ? formatPrice(Number(price)) : t("common.free")}</div>
          </CardContent>
        </Card>
      </div>
    </form>
    </FormProvider>
  );
}
