"use client";

import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { LoadingState } from "@/components/shared/states";
import { dashboardService } from "@/services/dashboard.service";
import { formatNumber } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

export default function StudioAnalyticsPage() {
  const t = useT();
  const { data: stats, isLoading } = useQuery({
    queryKey: ["studio", "teaching-stats"],
    queryFn: dashboardService.getTeachingStats,
  });

  if (isLoading || !stats) return <LoadingState className="min-h-[40vh]" />;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">{t("studio.analyticsTitle")}</h1>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">{t("studio.activeStudents")}</span>
          <div className="mt-2 text-3xl font-extrabold">{formatNumber(stats.activeStudents)}</div>
        </Card>
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">{t("studio.avgCompletion")}</span>
          <div className="mt-2 text-3xl font-extrabold">{stats.avgCompletion}%</div>
        </Card>
        <Card className="p-5">
          <span className="text-sm text-muted-foreground">{t("studio.avgQuizScore")}</span>
          <div className="mt-2 text-3xl font-extrabold">{Math.round(stats.avgQuizScore)}%</div>
        </Card>
      </div>

      <Card className="p-6">
        <h2 className="mb-4 text-lg font-bold">{t("studio.completionByCourse")}</h2>
        <div className="space-y-5">
          {stats.engagement.map((e) => (
            <div key={e.courseId}>
              <div className="mb-1 flex justify-between text-sm">
                <span className="font-medium">{e.title}</span>
                <span className="text-muted-foreground">{t("studio.completionMeta", { c: e.completion, n: e.students })}</span>
              </div>
              <Progress value={e.completion} />
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
