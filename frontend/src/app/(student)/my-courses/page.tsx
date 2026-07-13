"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, PlayCircle } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LoadingState, ErrorState, EmptyState } from "@/components/shared/states";
import { dashboardService } from "@/services/dashboard.service";
import { ROUTES } from "@/constants";
import { useT } from "@/providers/locale-provider";

type Filter = "all" | "in-progress" | "completed";

export default function MyCoursesPage() {
  const t = useT();
  const [filter, setFilter] = useState<Filter>("all");
  const enrolled = useQuery({ queryKey: ["my-courses"], queryFn: dashboardService.getEnrolled });

  const all = enrolled.data ?? [];
  const items = all.filter((e) =>
    filter === "all" ? true : filter === "completed" ? e.progress === 100 : e.progress < 100,
  );

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("mycourses.title")}</h1>
        <p className="text-muted-foreground">{t("mycourses.subtitle")}</p>
      </div>

      <Tabs value={filter} onValueChange={(v) => setFilter(v as Filter)}>
        <TabsList>
          <TabsTrigger value="all">{t("mycourses.all", { n: all.length })}</TabsTrigger>
          <TabsTrigger value="in-progress">{t("mycourses.inProgress")}</TabsTrigger>
          <TabsTrigger value="completed">{t("mycourses.completed")}</TabsTrigger>
        </TabsList>
      </Tabs>

      {enrolled.isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : enrolled.isError ? (
        <ErrorState onRetry={() => enrolled.refetch()} />
      ) : items.length === 0 ? (
        <EmptyState
          title={t("mycourses.emptyTitle")}
          description={t("mycourses.emptyDesc")}
          action={
            <Button asChild>
              <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
            </Button>
          }
        />
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((e) => {
            const done = e.progress === 100;
            return (
              <Card key={e.course.id} className="flex flex-col overflow-hidden">
                <div className={`relative aspect-video bg-gradient-to-br ${e.course.thumbnailColor}`}>
                  <span className="absolute inset-0 grid place-items-center text-white/90">
                    <PlayCircle className="size-12" />
                  </span>
                </div>
                <div className="flex flex-1 flex-col p-5">
                  <Badge variant={done ? "success" : "secondary"} className="w-fit">
                    {done ? t("mycourses.completedBadge") : t("mycourses.percentComplete", { p: e.progress })}
                  </Badge>
                  <h3 className="mt-2 line-clamp-2 font-bold">{e.course.title}</h3>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("mycourses.lessonsOf", { done: e.lessonsCompleted, total: e.course.totalLessons })}
                  </p>
                  <Progress value={e.progress} className="mt-3" />
                  <p className="mt-2 truncate text-xs text-muted-foreground">{t("mycourses.next", { lesson: e.currentLesson })}</p>
                  <div className="mt-auto flex gap-2 pt-4">
                    <Button asChild className="flex-1">
                      <Link href={ROUTES.learn(e.course.id)}>{done ? t("mycourses.review") : t("dash.resume")}</Link>
                    </Button>
                    {done && (
                      <Button asChild variant="outline" size="icon" aria-label="Certificate">
                        <Link href={ROUTES.certificates}>
                          <CheckCircle2 className="size-4" />
                        </Link>
                      </Button>
                    )}
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
