"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { ImagePlus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { coursesService } from "@/services/courses.service";
import { courseSchema, type CourseFormValues } from "./course-schema";
import { CATEGORIES, LANGUAGES, ROUTES } from "@/constants";
import { slugify } from "@/lib/utils";

export function CourseForm() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const {
    register,
    handleSubmit,
    control,
    watch,
    formState: { errors },
  } = useForm<CourseFormValues>({
    resolver: zodResolver(courseSchema),
    defaultValues: { lang: "uz", price: 0, isPublished: false, category: "" },
  });

  const title = watch("title") ?? "";
  const price = watch("price");
  const slug = slugify(title);

  const mutation = useMutation({
    mutationFn: (values: CourseFormValues) => coursesService.create(values),
    onSuccess: () => router.push(ROUTES.studioCourses),
    onError: (err) => setServerError(err instanceof Error ? err.message : "Failed to create course"),
  });

  const submit = (publish: boolean) =>
    handleSubmit((values) => {
      setServerError("");
      mutation.mutate({ ...values, isPublished: publish });
    });

  return (
    <form className="grid max-w-5xl gap-6 lg:grid-cols-3">
      <div className="space-y-6 lg:col-span-2">
        {serverError && <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{serverError}</div>}

        {/* Basic info */}
        <Card>
          <CardHeader>
            <CardTitle>Basic information</CardTitle>
            <CardDescription>This is what students see first.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-1.5">
              <Label htmlFor="title">
                Course title <span className="text-rose-500">*</span>
              </Label>
              <Input id="title" placeholder="e.g. Complete Next.js 16 Course" {...register("title")} />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{errors.title?.message ?? "Clear, specific titles convert best."}</span>
                <span>{title.length} / 200</span>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>URL slug (auto-generated)</Label>
              <div className="flex items-center overflow-hidden rounded-lg border bg-secondary/50">
                <span className="border-r bg-secondary px-3 py-2 text-sm text-muted-foreground">
                  learnhub.com/courses/
                </span>
                <span className="truncate px-3 py-2 text-sm text-muted-foreground">{slug || "your-course-title"}</span>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">
                Description <span className="text-rose-500">*</span>
              </Label>
              <Textarea id="description" rows={5} placeholder="What will students learn?" {...register("description")} />
              {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label>
                  Category <span className="text-rose-500">*</span>
                </Label>
                <Controller
                  control={control}
                  name="category"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select category" />
                      </SelectTrigger>
                      <SelectContent>
                        {CATEGORIES.map((c) => (
                          <SelectItem key={c} value={c}>
                            {c}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                {errors.category && <p className="text-xs text-destructive">{errors.category.message}</p>}
              </div>

              <div className="space-y-1.5">
                <Label>
                  Language <span className="text-rose-500">*</span>
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
          </CardContent>
        </Card>

        {/* Thumbnail */}
        <Card>
          <CardHeader>
            <CardTitle>Course thumbnail</CardTitle>
            <CardDescription>Recommended 1280×720 (16:9), max 2MB.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid place-items-center rounded-xl border-2 border-dashed p-8 text-center">
              <ImagePlus className="size-8 text-muted-foreground" />
              <p className="mt-2 text-sm font-semibold">
                Drag &amp; drop or <span className="text-primary">browse</span>
              </p>
              <p className="text-xs text-muted-foreground">PNG, JPG up to 2MB</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Sidebar: pricing + publish */}
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Pricing</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="price">Price (USD)</Label>
              <div className="flex items-center overflow-hidden rounded-lg border">
                <span className="border-r bg-secondary px-3 py-2 text-muted-foreground">$</span>
                <input
                  id="price"
                  type="number"
                  step="0.01"
                  className="flex-1 px-3 py-2 text-sm focus:outline-none"
                  {...register("price", { valueAsNumber: true })}
                />
              </div>
              <p className="text-xs text-muted-foreground">Set 0 to make this course free.</p>
              {errors.price && <p className="text-xs text-destructive">{errors.price.message}</p>}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Publish</CardTitle>
            <CardDescription>Drafts are only visible to you.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button type="button" className="w-full" disabled={mutation.isPending} onClick={submit(true)}>
              {mutation.isPending ? "Saving…" : "Publish course"}
            </Button>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={mutation.isPending}
              onClick={submit(false)}
            >
              Save as draft
            </Button>
          </CardContent>
        </Card>

        {/* Live preview */}
        <Card className="overflow-hidden">
          <div className="aspect-video bg-gradient-to-br from-indigo-400 to-violet-500" />
          <CardContent className="p-4">
            <h3 className="line-clamp-2 text-sm font-bold">{title || "Your course title"}</h3>
            <p className="mt-0.5 text-xs text-muted-foreground">by You</p>
            <div className="mt-2 font-extrabold">{Number(price) > 0 ? `$${Number(price).toFixed(2)}` : "Free"}</div>
          </CardContent>
        </Card>
      </div>
    </form>
  );
}
