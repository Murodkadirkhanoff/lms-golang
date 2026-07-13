"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { BookOpen, CheckCircle2, Clock, Award } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { LoadingState, ErrorState } from "@/components/shared/states";
import { dashboardService } from "@/services/dashboard.service";
import { ROUTES } from "@/constants";
import { formatDate } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

export default function DashboardPage() {
  const t = useT();
  const stats = useQuery({ queryKey: ["dashboard", "stats"], queryFn: dashboardService.getStats });
  const enrolled = useQuery({ queryKey: ["dashboard", "enrolled"], queryFn: dashboardService.getEnrolled });
  const certs = useQuery({ queryKey: ["dashboard", "certs"], queryFn: dashboardService.getCertificates });

  const statCards = [
    { label: t("dash.enrolled"), value: stats.data?.enrolled, icon: BookOpen, color: "text-primary" },
    { label: t("dash.inProgress"), value: stats.data?.inProgress, icon: Clock, color: "text-amber-600" },
    { label: t("dash.completed"), value: stats.data?.completed, icon: CheckCircle2, color: "text-emerald-600" },
    { label: t("dash.certificates"), value: stats.data?.certificates, icon: Award, color: "text-rose-600" },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("dash.welcome")}</h1>
        <p className="text-muted-foreground">{t("dash.streak")}</p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((s) => (
          <Card key={s.label} className="p-5">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{s.label}</span>
              <s.icon className={`size-5 ${s.color}`} />
            </div>
            <div className="mt-2 text-3xl font-extrabold">{s.value ?? "—"}</div>
          </Card>
        ))}
      </div>

      {/* Continue learning */}
      <Card className="p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-bold">{t("dash.continueLearning")}</h2>
        </div>
        {enrolled.isLoading ? (
          <LoadingState />
        ) : enrolled.isError ? (
          <ErrorState onRetry={() => enrolled.refetch()} />
        ) : (
          <div className="space-y-4">
            {enrolled.data?.map((e) => (
              <div key={e.course.id} className="flex items-center gap-4">
                <div className={`h-16 w-24 shrink-0 rounded-lg bg-gradient-to-br ${e.course.thumbnailColor}`} />
                <div className="min-w-0 flex-1">
                  <h3 className="truncate font-semibold">{e.course.title}</h3>
                  <p className="text-xs text-muted-foreground">
                    {t("dash.lessonsProgress", { done: e.lessonsCompleted, total: e.course.totalLessons, lesson: e.currentLesson })}
                  </p>
                  <Progress value={e.progress} className="mt-2" />
                </div>
                <Button asChild className="shrink-0">
                  <Link href={ROUTES.learn(e.course.id)}>{t("dash.resume")}</Link>
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* Recent certificates */}
      <Card className="p-6">
        <h2 className="mb-4 text-lg font-bold">{t("dash.recentCerts")}</h2>
        {certs.isLoading ? (
          <LoadingState />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {certs.data?.map((c) => (
              <div key={c.id} className={`flex items-center gap-3 rounded-xl p-3 ${c.color}`}>
                <Award className="size-6 text-amber-700" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold">{c.courseTitle}</div>
                  <div className="text-xs text-muted-foreground">{t("dash.issued", { date: formatDate(c.issuedAt) })}</div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}
