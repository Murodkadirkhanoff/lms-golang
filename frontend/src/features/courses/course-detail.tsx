"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  CheckCircle2,
  Check,
  Clock,
  FileText,
  Globe,
  Heart,
  Lock,
  Plus,
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
import { enrollmentsService } from "@/services/enrollments.service";
import { ROUTES } from "@/constants";
import { cn, formatNumber, formatPrice } from "@/lib/utils";
import { useCart } from "@/features/cart/cart-context";
import { useWishlist } from "@/features/wishlist/wishlist-context";
import { useAuth } from "@/providers/auth-provider";
import { useT } from "@/providers/locale-provider";

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
  const router = useRouter();
  const cart = useCart();
  const wishlist = useWishlist();
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const t = useT();

  // Bepul kursga yozilish — backendda enrollment yaratib learn'ga o'tadi.
  const enrollMutation = useMutation({
    mutationFn: (courseId: number) => enrollmentsService.enroll(courseId),
    onSuccess: (_, courseId) => {
      queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      queryClient.invalidateQueries({ queryKey: ["my-courses"] });
      router.push(ROUTES.learn(courseId));
    },
  });

  const enrollFree = (courseId: number) => {
    if (!isAuthenticated) {
      router.push(ROUTES.login);
      return;
    }
    enrollMutation.mutate(courseId);
  };

  if (isLoading) return <LoadingState className="min-h-[60vh]" />;
  if (isError || !course) return <ErrorState className="min-h-[60vh]" onRetry={() => refetch()} />;

  return (
    <>
      {/* Hero */}
      <section className="bg-slate-900 text-white">
        <div className="mx-auto max-w-7xl px-6 py-10">
          <nav className="mb-3 text-xs text-slate-400">{t("detail.home")} / {t(`cat.${course.category.replace(/\s/g, "")}`)}</nav>
          <div className="max-w-2xl">
            <h1 className="text-3xl font-extrabold leading-tight">{course.title}</h1>
            <p className="mt-3 text-slate-300">{course.description}</p>
            <div className="mt-4 flex flex-wrap items-center gap-4 text-sm">
              <span className="flex items-center gap-1 font-semibold text-amber-400">
                {course.rating} <Stars rating={course.rating} />
              </span>
              <span className="text-slate-300">
                ({t("detail.ratings", { ratings: formatNumber(course.ratingCount), students: formatNumber(course.studentCount) })})
              </span>
            </div>
            <div className="mt-3 flex items-center gap-2 text-sm text-slate-300">
              <div className={`size-7 rounded-full ${course.instructor.avatarColor}`} />
              {t("detail.createdBy")} <span className="font-semibold text-indigo-300">{course.instructor.name}</span>
            </div>
            <div className="mt-3 flex flex-wrap gap-4 text-xs text-slate-400">
              <span className="flex items-center gap-1">
                <Clock className="size-4" /> {course.totalDurationMinutes} {t("common.min")}
              </span>
              <span className="flex items-center gap-1">
                <Globe className="size-4" /> {course.lang.toUpperCase()}
              </span>
              <span className="flex items-center gap-1">
                <Award className="size-4" /> {t("detail.certificate")}
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
              <h2 className="text-xl font-bold">{t("detail.courseContent")}</h2>
              <span className="text-sm text-muted-foreground">
                {t("detail.contentMeta", { lessons: course.totalLessons, min: course.totalDurationMinutes })}
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
                      {t("detail.lessonsCount", { n: module.lessons.length })}
                      <ChevronDown className={`size-4 transition-transform ${openModule === idx ? "rotate-180" : ""}`} />
                    </span>
                  </button>
                  {openModule === idx && (
                    <ul className="text-sm">
                      {module.lessons.map((lesson) => {
                        const ownedByCourse = cart.hasCourse(course.id);
                        const inCart = cart.hasLesson(lesson.id);
                        const sellable = !lesson.isFree && lesson.price > 0 && !ownedByCourse;
                        return (
                          <li key={lesson.id} className="flex items-center justify-between gap-3 px-5 py-3 hover:bg-secondary/40">
                            <span className="flex min-w-0 items-center gap-3">
                              {lesson.isFree || ownedByCourse ? (
                                lesson.type === "text" ? (
                                  <FileText className="size-4 shrink-0 text-primary" />
                                ) : (
                                  <PlayCircle className="size-4 shrink-0 text-primary" />
                                )
                              ) : (
                                <Lock className="size-4 shrink-0 text-muted-foreground" />
                              )}
                              <span className="truncate">{lesson.title}</span>
                            </span>
                            <span className="flex shrink-0 items-center gap-3 text-xs text-muted-foreground">
                              {lesson.isFree && <span className="text-primary">{t("detail.preview")}</span>}
                              <span className="tabular-nums">{formatDuration(lesson.durationSeconds)}</span>
                              {ownedByCourse && !lesson.isFree && (
                                <span className="flex items-center gap-1 text-emerald-600">
                                  <Check className="size-3.5" /> {t("detail.lessonInCourse")}
                                </span>
                              )}
                              {sellable && (
                                <Button
                                  type="button"
                                  size="sm"
                                  variant={inCart ? "secondary" : "outline"}
                                  className="h-7 gap-1 px-2"
                                  onClick={() =>
                                    inCart ? cart.removeLesson(lesson.id) : cart.addLesson(course.id, lesson.id)
                                  }
                                >
                                  {inCart ? (
                                    <>
                                      <Check className="size-3.5" /> {t("detail.lessonAdded")}
                                    </>
                                  ) : (
                                    <>
                                      <Plus className="size-3.5" /> {formatPrice(lesson.price)}
                                    </>
                                  )}
                                </Button>
                              )}
                            </span>
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </div>
              ))}
              {(!course.modules || course.modules.length === 0) && (
                <p className="px-5 py-6 text-sm text-muted-foreground">{t("detail.curriculumSoon")}</p>
              )}
            </div>
          </section>

          {/* Instructor */}
          <Card className="p-6">
            <h2 className="mb-4 text-xl font-bold">{t("detail.instructor")}</h2>
            <div className="flex gap-4">
              <Link href={ROUTES.instructor(course.instructor.id)} className={`size-20 shrink-0 rounded-full ${course.instructor.avatarColor}`} />
              <div>
                <Link href={ROUTES.instructor(course.instructor.id)} className="font-bold text-primary hover:underline">
                  {course.instructor.name}
                </Link>
                <p className="text-sm text-muted-foreground">{course.instructor.headline}</p>
                <div className="mt-2 flex flex-wrap gap-4 text-xs text-muted-foreground">
                  <span>★ {course.instructor.rating} {t("detail.ratingShort")}</span>
                  <span>👥 {formatNumber(course.instructor.students)} {t("common.students")}</span>
                  <span>📚 {course.instructor.courses} {t("detail.coursesCount")}</span>
                </div>
              </div>
            </div>
          </Card>

          {/* Reviews */}
          {course.reviews && course.reviews.length > 0 && (
            <Card className="p-6">
              <h2 className="mb-4 text-xl font-bold">{t("detail.studentReviews")}</h2>
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
                {course.price === 0 ? (
                  <>
                    <Button
                      className="mt-4 w-full"
                      size="lg"
                      disabled={enrollMutation.isPending}
                      onClick={() => enrollFree(course.id)}
                    >
                      {t("detail.enrollFree")}
                    </Button>
                    {enrollMutation.isError && (
                      <p className="mt-2 text-center text-xs text-destructive">
                        {enrollMutation.error instanceof Error ? enrollMutation.error.message : "Failed to enroll"}
                      </p>
                    )}
                  </>
                ) : (
                  <>
                    <Button
                      className="mt-4 w-full"
                      size="lg"
                      onClick={() => {
                        cart.addCourse(course.id);
                        router.push(ROUTES.checkout);
                      }}
                    >
                      {t("detail.buyNow")}
                    </Button>
                    <Button
                      variant="outline"
                      className="mt-2 w-full"
                      size="lg"
                      onClick={() => {
                        if (cart.hasCourse(course.id)) {
                          router.push(ROUTES.cart);
                        } else {
                          cart.addCourse(course.id);
                        }
                      }}
                    >
                      {cart.hasCourse(course.id) ? t("detail.goToCart") : t("detail.addToCart")}
                    </Button>
                    <p className="mt-3 text-center text-xs text-muted-foreground">{t("detail.orBuyLessons")}</p>
                  </>
                )}
                <Button
                  variant="ghost"
                  className="mt-2 w-full"
                  size="lg"
                  onClick={() => wishlist.toggle(course.id)}
                >
                  <Heart className={cn("size-4", wishlist.has(course.id) && "fill-rose-500 text-rose-500")} />
                  {wishlist.has(course.id) ? t("detail.savedWishlist") : t("detail.addWishlist")}
                </Button>
                <p className="mt-3 text-center text-xs text-muted-foreground">{t("detail.guarantee")}</p>
                <div className="mt-5 space-y-2 border-t pt-5 text-sm text-muted-foreground">
                  <p className="font-semibold text-foreground">{t("detail.includes")}</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("detail.incVideo", { min: course.totalDurationMinutes })}</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("detail.incLessons", { n: course.totalLessons })}</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("detail.incAccess")}</p>
                  <p className="flex items-center gap-2"><CheckCircle2 className="size-4 text-emerald-600" /> {t("detail.incCert")}</p>
                </div>
              </div>
            </Card>
          </div>
        </aside>
      </div>
    </>
  );
}
