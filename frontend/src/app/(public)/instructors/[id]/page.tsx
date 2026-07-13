"use client";

import { use } from "react";
import { useQuery } from "@tanstack/react-query";
import { Award, BookOpen, MessageSquare, Star, Users } from "lucide-react";
import { Card } from "@/components/ui/card";
import { CourseCard } from "@/features/courses/course-card";
import { LoadingState, ErrorState, CardGridSkeleton } from "@/components/shared/states";
import { instructorsService } from "@/services/instructors.service";
import { coursesService } from "@/services/courses.service";
import { formatNumber } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

export default function InstructorPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const t = useT();

  const instructor = useQuery({
    queryKey: ["instructor", id],
    queryFn: () => instructorsService.getById(id),
  });
  const courses = useQuery({
    queryKey: ["instructor-courses", id],
    queryFn: () => coursesService.getByInstructor(Number(id)),
  });

  if (instructor.isLoading) return <LoadingState className="min-h-[60vh]" />;
  if (instructor.isError || !instructor.data)
    return <ErrorState className="min-h-[60vh]" onRetry={() => instructor.refetch()} />;

  const i = instructor.data;
  const stats = [
    { label: t("instructor.rating"), value: i.rating, icon: Star, color: "text-amber-500" },
    { label: t("instructor.students"), value: formatNumber(i.students), icon: Users, color: "text-primary" },
    { label: t("instructor.courses"), value: i.courses, icon: BookOpen, color: "text-emerald-600" },
    { label: t("instructor.reviews"), value: formatNumber(Math.round(i.students * 0.18)), icon: Award, color: "text-rose-600" },
  ];

  return (
    <>
      <section className="bg-slate-900 text-white">
        <div className="mx-auto flex max-w-7xl flex-col gap-6 px-6 py-12 sm:flex-row sm:items-center">
          <div className={`size-28 shrink-0 rounded-full ${i.avatarColor} ring-4 ring-white/10`} />
          <div>
            <p className="text-sm font-semibold uppercase tracking-wide text-indigo-300">{t("instructor.label")}</p>
            <h1 className="mt-1 text-3xl font-extrabold">{i.name}</h1>
            <p className="mt-1 text-slate-300">{i.headline}</p>
            <div className="mt-4 flex flex-wrap gap-6 text-sm">
              {stats.map((s) => (
                <div key={s.label} className="flex items-center gap-2">
                  <s.icon className={`size-4 ${s.color}`} />
                  <span className="font-bold">{s.value}</span>
                  <span className="text-slate-400">{s.label}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-7xl gap-10 px-6 py-10 lg:grid-cols-3">
        <aside className="space-y-6">
          <Card className="p-6">
            <h2 className="font-bold">{t("instructor.about")}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {t("instructor.bio", { name: i.name, courses: i.courses, students: formatNumber(i.students) })}
            </p>
            <button className="mt-4 flex items-center gap-2 text-sm font-semibold text-primary hover:underline">
              <MessageSquare className="size-4" /> {t("instructor.sendMessage")}
            </button>
          </Card>
        </aside>

        <div className="lg:col-span-2">
          <h2 className="mb-6 text-xl font-bold">{t("instructor.coursesBy", { name: i.name })}</h2>
          {courses.isLoading ? (
            <CardGridSkeleton />
          ) : (
            <div className="grid gap-6 sm:grid-cols-2">
              {(courses.data ?? []).map((c) => (
                <CourseCard key={c.id} course={c} />
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
