"use client";

import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { coursesService } from "@/services/courses.service";
import { formatNumber, formatPrice } from "@/lib/utils";

const kpis = [
  { label: "Total revenue", value: "$48,250", trend: "▲ 12.5% vs last month", trendColor: "text-emerald-600" },
  { label: "Total students", value: "12,840", trend: "▲ 320 new this week", trendColor: "text-emerald-600" },
  { label: "Published courses", value: "14", trend: "2 in draft", trendColor: "text-muted-foreground" },
  { label: "Avg. rating", value: "4.8 ★", trend: "from 9,200 reviews", trendColor: "text-muted-foreground" },
];

const revenueBars = [45, 60, 52, 78, 68, 100];
const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun"];

export default function StudioDashboardPage() {
  const { data: top } = useQuery({ queryKey: ["studio", "top"], queryFn: () => coursesService.getPopular(4) });

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">Dashboard</h1>

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
          <h2 className="mb-6 text-lg font-bold">Revenue overview</h2>
          <div className="flex h-52 items-end justify-between gap-3">
            {revenueBars.map((h, i) => (
              <div key={i} className="flex flex-1 flex-col items-center gap-2">
                <div className="w-full rounded-t bg-primary/80" style={{ height: `${h}%` }} />
                <span className="text-xs text-muted-foreground">{months[i]}</span>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6">
          <h2 className="mb-4 text-lg font-bold">Top courses</h2>
          <div className="space-y-4">
            {top?.map((c) => (
              <div key={c.id} className="flex items-center gap-3">
                <div className={`size-10 rounded-lg bg-gradient-to-br ${c.thumbnailColor}`} />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-semibold">{c.title}</div>
                  <div className="text-xs text-muted-foreground">{formatNumber(c.studentCount)} students</div>
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
