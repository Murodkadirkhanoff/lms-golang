"use client";

import { useQuery } from "@tanstack/react-query";
import { Card } from "@/components/ui/card";
import { adminService } from "@/services/admin.service";
import { formatNumber, formatPrice } from "@/lib/utils";

const growth = [
  [50, 40],
  [62, 55],
  [70, 60],
  [80, 72],
  [92, 85],
  [100, 95],
];

export default function AdminOverviewPage() {
  const { data } = useQuery({ queryKey: ["admin", "stats"], queryFn: adminService.getStats });

  const kpis = [
    { label: "Total users", value: data ? formatNumber(data.totalUsers) : "—", trend: "▲ 4,120 this week" },
    { label: "Total courses", value: data ? formatNumber(data.totalCourses) : "—", trend: "18 pending review" },
    { label: "Platform revenue", value: data ? formatPrice(data.revenue) : "—", trend: "▲ 8.3% MoM" },
    { label: "Active instructors", value: data ? formatNumber(data.activeInstructors) : "—", trend: "62 new applications" },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-extrabold">Platform overview</h1>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {kpis.map((k) => (
          <Card key={k.label} className="p-5">
            <span className="text-sm text-muted-foreground">{k.label}</span>
            <div className="mt-2 text-3xl font-extrabold">{k.value}</div>
            <div className="mt-1 text-xs text-emerald-600">{k.trend}</div>
          </Card>
        ))}
      </div>

      <Card className="p-6">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-lg font-bold">User &amp; revenue growth</h2>
          <div className="flex gap-3 text-xs">
            <span className="flex items-center gap-1">
              <span className="size-2 rounded-full bg-rose-500" /> Users
            </span>
            <span className="flex items-center gap-1">
              <span className="size-2 rounded-full bg-primary" /> Revenue
            </span>
          </div>
        </div>
        <div className="flex h-52 items-end justify-between gap-2">
          {growth.map(([u, r], i) => (
            <div key={i} className="flex flex-1 items-end gap-1">
              <div className="w-1/2 rounded-t bg-rose-400" style={{ height: `${u}%` }} />
              <div className="w-1/2 rounded-t bg-primary" style={{ height: `${r}%` }} />
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
