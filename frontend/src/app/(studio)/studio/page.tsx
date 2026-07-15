"use client";

import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { coursesService } from "@/services/courses.service";
import { dashboardService } from "@/services/dashboard.service";
import { formatNumber, formatPrice } from "@/lib/utils";
import { useAuth } from "@/providers/auth-provider";
import { useT } from "@/providers/locale-provider";

// "YYYY-MM" -> "Feb" ko'rinishidagi qisqa oy nomi.
function monthLabel(month: string): string {
  const date = new Date(`${month}-01T00:00:00Z`);
  return date.toLocaleString("en", { month: "short", timeZone: "UTC" });
}

export default function StudioDashboardPage() {
  const t = useT();
  const { user } = useAuth();

  const { data: stats } = useQuery({
    queryKey: ["studio", "teaching-stats"],
    queryFn: dashboardService.getTeachingStats,
  });

  const { data: myCourses } = useQuery({
    queryKey: ["studio", "my-courses", user?.id],
    queryFn: () => coursesService.getByInstructor(user!.id),
    enabled: user != null,
  });

  const top = [...(myCourses ?? [])].sort((a, b) => b.studentCount - a.studentCount).slice(0, 4);

  const kpis = [
    {
      label: t("studio.totalRevenue"),
      value: stats ? formatPrice(stats.totalRevenue) : "—",
      trend: t("studio.revVsMonth"),
      trendColor: "text-emerald-600",
    },
    {
      label: t("studio.totalStudents"),
      value: stats ? formatNumber(stats.totalStudents) : "—",
      trend: t("studio.newThisWeek"),
      trendColor: "text-emerald-600",
    },
    {
      label: t("studio.publishedCourses"),
      value: stats ? String(stats.publishedCourses) : "—",
      trend: t("studio.inDraft"),
      trendColor: "text-muted-foreground",
    },
    {
      label: t("studio.avgRating"),
      value: stats && stats.avgRating > 0 ? `${stats.avgRating.toFixed(1)} ★` : "—",
      trend: t("studio.fromReviews"),
      trendColor: "text-muted-foreground",
    },
  ];

  const monthly = stats?.monthlyRevenue ?? [];
  const maxRevenue = Math.max(...monthly.map((m) => m.revenue), 1);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">{t("studio.dashboard")}</h1>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {kpis.map((k) => (
          <Card key={k.label} className="p-5">
            <span className="text-sm text-muted-foreground">{k.label}</span>
            <div className="mt-2 text-3xl font-extrabold">{k.value}</div>
            <div className={`mt-1 text-xs ${k.trendColor}`}>{k.trend}</div>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="p-6 lg:col-span-2">
          <h2 className="mb-6 text-lg font-bold">{t("studio.revenueOverview")}</h2>
          <div className="flex h-52 items-end justify-between gap-3">
            {monthly.map((m) => (
              <div key={m.month} className="flex flex-1 flex-col items-center gap-2">
                <div
                  className="w-full rounded-t bg-primary/80"
                  style={{ height: `${Math.round((m.revenue / maxRevenue) * 100)}%` }}
                  title={formatPrice(m.revenue)}
                />
                <span className="text-xs text-muted-foreground">{monthLabel(m.month)}</span>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6">
          <h2 className="mb-4 text-lg font-bold">{t("studio.topCourses")}</h2>
          <div className="space-y-4">
            {top.map((c) => (
              <div key={c.id} className="flex items-center gap-3">
                <div className={`size-10 rounded-lg bg-gradient-to-br ${c.thumbnailColor}`} />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold">{c.title}</div>
                  <div className="text-xs text-muted-foreground">{t("studio.studentsCount", { n: formatNumber(c.studentCount) })}</div>
                </div>
                <span className="text-sm font-bold">{formatPrice(c.price)}</span>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}
