"use client";

import { use } from "react";
import { useQuery } from "@tanstack/react-query";
import { CourseForm } from "@/features/courses/course-form";
import { coursesService } from "@/services/courses.service";
import { useT } from "@/providers/locale-provider";

export default function EditCoursePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const t = useT();

  const {
    data: course,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["course", id],
    queryFn: () => coursesService.getById(id),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("studio.editCourseTitle")}</h1>
        <p className="text-muted-foreground">{t("studio.courseHash", { id })}</p>
      </div>
      {isLoading && (
        <div className="grid gap-4">
          <div className="h-40 animate-pulse rounded-xl bg-secondary" />
          <div className="h-64 animate-pulse rounded-xl bg-secondary" />
        </div>
      )}
      {error && (
        <div className="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {error instanceof Error ? error.message : t("common.error")}
        </div>
      )}
      {/* key: kurs yuklangach forma defaultValues bilan qayta quriladi */}
      {course && <CourseForm key={course.id} course={course} />}
    </div>
  );
}
