"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  CheckCircle2,
  Clock,
  Globe,
  Lock,
  PlayCircle,
  Award,
  ChevronDown,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Stars } from "@/components/shared/stars";
import { LoadingState, ErrorState } from "@/components/shared/states";
import { coursesService } from "@/services/courses.service";
import { formatNumber, formatPrice } from "@/lib/utils";

function formatDuration(seconds: number) {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function CourseDetail({ slug }: { slug: string }) {
  const { data: course, isLoading, isError, refetch } = useQuery({
    queryKey: ["course", slug],
    queryFn: () => coursesService.getBySlug(slug),
  });
  const [openModule, setOpenModule] = useState(0);

  if (isLoading) return <LoadingState className="min-h-[60vh]" />;
  if (isError || !course) return <ErrorState className="min-h-[60vh]" onRetry={() => refetch()} />;

  return (
    <>
      {/* Hero */}
      <section className="bg-slate-900 text-white">
        <div className="mx-auto max-w-7xl px-6 py-10">
          <nav className="mb-3 text-xs text-slate-400">Home / {course.category}</nav>
          <div className="max-w-2xl">
            <h1 className="text-3xl font-extrabold leading-tight">{course.title}</h1>
            <p className="mt-3 text-slate-300">{course.description}</p>
            <div className="mt-4 flex flex-wrap items-center gap-4 text-sm">
              <span className="flex items-center gap-1 font-semibold text-amber-400">
                {course.rating} <Stars rating={course.rating} />
              </span>
              <span className="text-slate-300">
                ({formatNumber(course.ratingCount)} ratings · {formatNumber(course.studentCount)} students)
              </span>
            </div>
            <div className="mt-3 flex items-center gap-2 text-sm text-slate-300">
              <div className={`size-7 rounded-full ${course.instructor.avatarColor}`} />
              Created by <span className="font-semibold text-indigo-300">{course.instructor.name}</span>
            </div>
            <div className="mt-3 flex flex-wrap gap-4 text-xs text-slate-400">
              <span className="flex items-center gap-1">
                <Clock className="size-4" /> {course.totalDurationMinutes} min
              </span>
              <span className="flex items-center gap-1">
                <Globe className="size-4" /> {course.lang.toUpperCase()}
              </span>
              <span className="flex items-center gap-1">
                <Award className="size-4" /> Certificate
              </span>
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-7xl gap-10 px-6 py-10 lg:grid-cols-3">
        {/* Left */}
        <div className="space-y-10 lg:col-span-2">
          {/* Curriculum */}
          <section>
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-xl font-bold">Course content</h2>
              <span className="text-sm text-muted-foreground">
                {course.totalLessons} lessons · {course.totalDurationMinutes} min
              </span>
            </div>
            <div className="divide-y overflow-hidden rounded-xl border">
              {(course.modules ?? []).map((module, idx) => (
                <div key={module.id}>
                  <button
                    onClick={() => setOpenModule(openModule === idx ? -1 : idx)}
                    className="flex w-full items-center justify-between bg-secondary/50 px-5 py-4 text-left font-semibold"
                  >
                    <span>{module.title}</span>
                    <span className="flex items-center gap-2 text-sm font-normal text-muted-foreground">
                      {module.lessons.length} lessons
                      <ChevronDown className={`size-4 transition-transform ${openModule === idx ? "rotate-180" : ""}`} />
                    </span>
                  </button>
                  {openModule === idx && (
                    <ul className="text-sm">
                      {module.lessons.map((lesson) => (
                        <li key={lesson.id} className="flex items-center justify-between px-5 py-3 hover:bg-secondary/40">
                          <span className="flex items-center gap-3">
                            {lesson.isFree ? (
                              <PlayCircle className="size-4 text-primary" />
                            ) : (
                              <Lock className="size-4 text-muted-foreground" />
                            )}
                            {lesson.title}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {lesson.isFree && <span className="mr-2 text-primary">Preview</span>}
                            {formatDuration(lesson.durationSeconds)}
                          </span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              ))}
              {(!course.modules || course.modules.length === 0) && (
                <p className="px-5 py-6 text-sm text-muted-foreground">Curriculum coming soon.</p>
              )}
            </div>
          </section>

          {/* Instructor */}
          <Card className="p-6">
            <h2 className="mb-4 text-xl font-bold">Instructor</h2>
            <div className="flex gap-4">
              <div className={`size-20 shrink-0 rounded-full ${course.instructor.avatarColor}`} />
              <div>
                <h3 className="font-bold text-primary">{course.instructor.name}</h3>
                <p className="text-sm text-muted-foreground">{course.instructor.headline}</p>
                <div className="mt-2 flex flex-wrap gap-4 text-xs text-muted-foreground">
                  <span>★ {course.instructor.rating} rating</span>
                  <span>👥 {formatNumber(course.instructor.students)} students</span>
                  <span>📚 {course.instructor.courses} courses</span>
                </div>
              </div>
            </div>
          </Card>

          {/* Reviews */}
          {course.reviews && course.reviews.length > 0 && (
            <Card className="p-6">
              <h2 className="mb-4 text-xl font-bold">Student reviews</h2>
              <div className="space-y-5">
                {course.reviews.map((review) => (
                  <div key={review.id} className="border-t pt-5 first:border-t-0 first:pt-0">
                    <div className="flex items-center gap-3">
                      <div className={`size-9 rounded-full ${review.avatarColor}`} />
                      <div>
                        <div className="text-sm font-semibold">{review.user}</div>
                        <Stars rating={review.rating} />
                      </div>
                    </div>
                    <p className="mt-2 text-sm text-muted-foreground">{review.comment}</p>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </div>

        {/* Sticky purchase card */}
        <aside>
          <div className="lg:sticky lg:top-24">
            <Card className="overflow-hidden shadow-lg">
              <div className={`grid aspect-video place-items-center bg-gradient-to-br ${course.thumbnailColor} text-white`}>
                <PlayCircle className="size-14" />
              </div>
              <div className="p-6">
                <div className="flex items-end gap-2">
                  <span className="text-3xl font-extrabold">{formatPrice(course.price)}</span>
                </div>
                <Button className="mt-4 w-full" size="lg">
                  {course.price === 0 ? "Enroll for free" : "Enroll now"}
                </Button>
                <Button variant="outline" className="mt-2 w-full" size="lg">
                  Add to cart
                </Button>
                <p className="mt-3 text-center text-xs text-muted-foreground">30-day money-back guarantee</p>
                <div className="mt-5 space-y-2 border-t pt-5 text-sm text-muted-foreground">
                  <p className="font-semibold text-foreground">This course includes:</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {course.totalDurationMinutes} min on-demand video</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {course.totalLessons} lessons</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> Full lifetime access</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> Certificate of completion</p>
                </div>
              </div>
            </Card>
          </div>
        </aside>
      </div>
    </>
  );
}
