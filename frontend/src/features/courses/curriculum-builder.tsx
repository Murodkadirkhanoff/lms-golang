"use client";

import { useRef } from "react";
import { Controller, useFieldArray, useFormContext } from "react-hook-form";
import { FileText, GripVertical, Plus, Trash2, UploadCloud, Video } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { EmptyState } from "@/components/shared/states";
import { cn } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";
import type { CourseFormValues } from "./course-schema";

export function CurriculumBuilder() {
  const t = useT();
  const { control } = useFormContext<CourseFormValues>();
  const { fields, append, remove } = useFieldArray({ control, name: "modules" });

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("cb.title")}</CardTitle>
        <CardDescription>{t("cb.desc")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {fields.length === 0 ? (
          <EmptyState title={t("cb.noSections")} description={t("cb.noSectionsDesc")} />
        ) : (
          fields.map((field, index) => (
            <ModuleEditor key={field.id} index={index} onRemove={() => remove(index)} />
          ))
        )}

        <Button
          type="button"
          variant="outline"
          className="w-full border-dashed"
          onClick={() => append({ title: "", lessons: [] })}
        >
          <Plus className="size-4" /> {t("cb.addSection")}
        </Button>
      </CardContent>
    </Card>
  );
}

function ModuleEditor({ index, onRemove }: { index: number; onRemove: () => void }) {
  const t = useT();
  const {
    control,
    register,
    formState: { errors },
  } = useFormContext<CourseFormValues>();
  const { fields, append, remove } = useFieldArray({ control, name: `modules.${index}.lessons` });
  const titleError = errors.modules?.[index]?.title?.message;

  return (
    <div className="rounded-xl border">
      {/* Section header */}
      <div className="flex items-center gap-2 border-b bg-secondary/50 p-3">
        <GripVertical className="size-4 shrink-0 text-muted-foreground" />
        <span className="shrink-0 text-sm font-semibold text-muted-foreground">{t("cb.section", { n: index + 1 })}</span>
        <Input
          placeholder={t("cb.sectionPlaceholder")}
          className="h-9 bg-background"
          {...register(`modules.${index}.title`)}
        />
        <Button type="button" variant="ghost" size="icon" className="shrink-0 text-rose-600" onClick={onRemove}>
          <Trash2 className="size-4" />
        </Button>
      </div>
      {titleError && <p className="px-3 pt-2 text-xs text-destructive">{titleError}</p>}

      {/* Lessons */}
      <div className="space-y-3 p-3">
        {fields.map((field, li) => (
          <LessonRow
            key={field.id}
            moduleIndex={index}
            lessonIndex={li}
            onRemove={() => remove(li)}
          />
        ))}

        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-primary"
          onClick={() =>
            append({ title: "", type: "video", contentUrl: "", content: "", durationMinutes: 0, price: 0, isFree: false })
          }
        >
          <Plus className="size-4" /> {t("cb.addLesson")}
        </Button>
      </div>
    </div>
  );
}

function LessonRow({
  moduleIndex,
  lessonIndex,
  onRemove,
}: {
  moduleIndex: number;
  lessonIndex: number;
  onRemove: () => void;
}) {
  const t = useT();
  const {
    control,
    register,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<CourseFormValues>();
  const fileInput = useRef<HTMLInputElement>(null);

  const base = `modules.${moduleIndex}.lessons.${lessonIndex}` as const;
  const isFree = watch(`${base}.isFree`);
  const type = watch(`${base}.type`);
  const contentUrl = watch(`${base}.contentUrl`);
  const lessonErrors = errors.modules?.[moduleIndex]?.lessons?.[lessonIndex];

  const LESSON_TYPES = [
    { value: "video", label: t("cb.typeVideo"), icon: Video },
    { value: "text", label: t("cb.typeText"), icon: FileText },
  ] as const;

  return (
    <div className="rounded-lg border bg-background p-3">
      <div className="flex items-start gap-2">
        <GripVertical className="mt-2.5 size-4 shrink-0 text-muted-foreground" />
        <div className="flex-1 space-y-3">
          {/* Title */}
          <div>
            <Input placeholder={t("cb.lessonTitle")} className="h-9" {...register(`${base}.title`)} />
            {lessonErrors?.title && <p className="mt-1 text-xs text-destructive">{lessonErrors.title.message}</p>}
          </div>

          {/* Lesson type selector */}
          <div className="flex gap-2">
            {LESSON_TYPES.map((lt) => (
              <button
                key={lt.value}
                type="button"
                onClick={() => setValue(`${base}.type`, lt.value)}
                className={cn(
                  "flex flex-1 items-center justify-center gap-2 rounded-lg border py-2 text-sm font-medium transition-colors",
                  type === lt.value ? "border-primary bg-accent text-accent-foreground" : "hover:bg-secondary/50",
                )}
              >
                <lt.icon className="size-4" /> {lt.label}
              </button>
            ))}
          </div>

          {/* Content: video upload or markdown editor */}
          {type === "video" ? (
            <div className="space-y-2">
              <button
                type="button"
                onClick={() => fileInput.current?.click()}
                className="flex w-full flex-col items-center gap-1 rounded-lg border-2 border-dashed p-4 text-center hover:bg-secondary/40"
              >
                <UploadCloud className="size-6 text-muted-foreground" />
                <span className="max-w-full break-all text-sm font-medium">
                  {contentUrl ? contentUrl : t("cb.uploadVideo")}
                </span>
                <span className="text-xs text-muted-foreground">{t("cb.uploadHint")}</span>
              </button>
              <input
                ref={fileInput}
                type="file"
                accept="video/*"
                className="hidden"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) setValue(`${base}.contentUrl`, file.name, { shouldValidate: true });
                }}
              />
              {lessonErrors?.contentUrl && (
                <p className="text-xs text-destructive">{lessonErrors.contentUrl.message}</p>
              )}
            </div>
          ) : (
            <div>
              <Controller
                control={control}
                name={`${base}.content`}
                render={({ field }) => (
                  <MarkdownEditor value={field.value ?? ""} onChange={field.onChange} />
                )}
              />
              {lessonErrors?.content && <p className="mt-1 text-xs text-destructive">{lessonErrors.content.message}</p>}
            </div>
          )}

          {/* Duration (video) + price */}
          <div className="grid gap-2 sm:grid-cols-2">
            {type === "video" && (
              <div className="flex items-center overflow-hidden rounded-lg border">
                <input
                  type="number"
                  min={0}
                  placeholder="0"
                  className="h-9 w-full px-3 text-sm focus:outline-none"
                  {...register(`${base}.durationMinutes`, { valueAsNumber: true })}
                />
                <span className="border-l bg-secondary px-3 py-2 text-xs text-muted-foreground">{t("common.min")}</span>
              </div>
            )}
            <div>
              <div className="flex items-center overflow-hidden rounded-lg border">
                <input
                  type="number"
                  step="1000"
                  min={0}
                  disabled={isFree}
                  placeholder="0"
                  className="h-9 w-full px-3 text-sm focus:outline-none disabled:bg-secondary/50 disabled:text-muted-foreground"
                  {...register(`${base}.price`, { valueAsNumber: true })}
                />
                <span className="border-l bg-secondary px-3 py-2 text-xs text-muted-foreground">{t("common.sum")}</span>
              </div>
              {lessonErrors?.price && <p className="mt-1 text-xs text-destructive">{lessonErrors.price.message}</p>}
            </div>
          </div>

          {/* Free toggle (enforces is_free = false OR price = 0) */}
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              className="size-4 accent-[var(--primary)]"
              {...register(`${base}.isFree`)}
              onChange={(e) => {
                setValue(`${base}.isFree`, e.target.checked);
                if (e.target.checked) setValue(`${base}.price`, 0);
              }}
            />
            <span className="text-muted-foreground">{t("cb.freePreview")}</span>
          </label>
        </div>

        <Button type="button" variant="ghost" size="icon" className="shrink-0 text-rose-600" onClick={onRemove}>
          <Trash2 className="size-4" />
        </Button>
      </div>
    </div>
  );
}
