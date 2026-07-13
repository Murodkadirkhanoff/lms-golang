"use client";

import { CourseForm } from "@/features/courses/course-form";
import { useT } from "@/providers/locale-provider";

export default function NewCoursePage() {
  const t = useT();
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("studio.newCourseTitle")}</h1>
        <p className="text-muted-foreground">{t("studio.newCourseSubtitle")}</p>
      </div>
      <CourseForm />
    </div>
  );
}
