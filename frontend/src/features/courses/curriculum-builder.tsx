"use client";

import { useFieldArray, useFormContext } from "react-hook-form";
import { GripVertical, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/shared/states";
import type { CourseFormValues } from "./course-schema";

export function CurriculumBuilder() {
  const { control } = useFormContext<CourseFormValues>();
  const { fields, append, remove } = useFieldArray({ control, name: "modules" });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Curriculum</CardTitle>
        <CardDescription>Organize your course into sections (modules) and lessons.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {fields.length === 0 ? (
          <EmptyState
            title="No sections yet"
            description="Add your first section to start building the curriculum."
          />
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
          <Plus className="size-4" /> Add section
        </Button>
      </CardContent>
    </Card>
  );
}

function ModuleEditor({ index, onRemove }: { index: number; onRemove: () => void }) {
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
        <span className="shrink-0 text-sm font-semibold text-muted-foreground">Section {index + 1}</span>
        <Input
          placeholder="e.g. Getting Started"
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
            append({ title: "", contentUrl: "", durationMinutes: 0, price: 0, isFree: false })
          }
        >
          <Plus className="size-4" /> Add lesson
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
  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<CourseFormValues>();

  const base = `modules.${moduleIndex}.lessons.${lessonIndex}` as const;
  const isFree = watch(`${base}.isFree`);
  const lessonErrors = errors.modules?.[moduleIndex]?.lessons?.[lessonIndex];

  return (
    <div className="rounded-lg border bg-background p-3">
      <div className="flex items-start gap-2">
        <GripVertical className="mt-2.5 size-4 shrink-0 text-muted-foreground" />
        <div className="grid flex-1 gap-2 sm:grid-cols-2">
          {/* Title */}
          <div className="sm:col-span-2">
            <Input placeholder="Lesson title" className="h-9" {...register(`${base}.title`)} />
            {lessonErrors?.title && <p className="mt-1 text-xs text-destructive">{lessonErrors.title.message}</p>}
          </div>

          {/* Content URL */}
          <div className="sm:col-span-2">
            <Input placeholder="Video / content URL (https://…)" className="h-9" {...register(`${base}.contentUrl`)} />
          </div>

          {/* Duration */}
          <div>
            <div className="flex items-center overflow-hidden rounded-lg border">
              <input
                type="number"
                min={0}
                placeholder="0"
                className="h-9 w-full px-3 text-sm focus:outline-none"
                {...register(`${base}.durationMinutes`, { valueAsNumber: true })}
              />
              <span className="border-l bg-secondary px-3 py-2 text-xs text-muted-foreground">min</span>
            </div>
          </div>

          {/* Price */}
          <div>
            <div className="flex items-center overflow-hidden rounded-lg border">
              <span className="border-r bg-secondary px-3 py-2 text-xs text-muted-foreground">$</span>
              <input
                type="number"
                step="0.01"
                min={0}
                disabled={isFree}
                placeholder="0.00"
                className="h-9 w-full px-3 text-sm focus:outline-none disabled:bg-secondary/50 disabled:text-muted-foreground"
                {...register(`${base}.price`, { valueAsNumber: true })}
              />
            </div>
            {lessonErrors?.price && <p className="mt-1 text-xs text-destructive">{lessonErrors.price.message}</p>}
          </div>

          {/* Free toggle (enforces is_free = false OR price = 0) */}
          <label className="flex items-center gap-2 text-sm sm:col-span-2">
            <input
              type="checkbox"
              className="size-4 accent-[var(--primary)]"
              {...register(`${base}.isFree`)}
              onChange={(e) => {
                setValue(`${base}.isFree`, e.target.checked);
                if (e.target.checked) setValue(`${base}.price`, 0);
              }}
            />
            <span className="text-muted-foreground">Free preview lesson</span>
          </label>
        </div>

        <Button type="button" variant="ghost" size="icon" className="shrink-0 text-rose-600" onClick={onRemove}>
          <Trash2 className="size-4" />
        </Button>
      </div>
    </div>
  );
}
