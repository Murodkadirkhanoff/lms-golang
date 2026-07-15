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
      <CourseForm courseId={Number(id)} />
    </div>
  );
}
