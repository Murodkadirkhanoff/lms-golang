"use client";

import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Bell, BookOpen, CheckCheck, Megaphone, MessageSquare, Settings } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LoadingState, ErrorState, EmptyState } from "@/components/shared/states";
import { notificationsService } from "@/services/notifications.service";
import { cn } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";
import type { Notification, NotificationType } from "@/types";

const ICONS: Record<NotificationType, { icon: LucideIcon; color: string }> = {
  course: { icon: BookOpen, color: "bg-indigo-100 text-indigo-600" },
  message: { icon: MessageSquare, color: "bg-emerald-100 text-emerald-600" },
  promo: { icon: Megaphone, color: "bg-amber-100 text-amber-600" },
  system: { icon: Settings, color: "bg-slate-100 text-slate-600" },
};

export default function NotificationsPage() {
  const t = useT();
  const query = useQuery({ queryKey: ["notifications"], queryFn: notificationsService.list });
  const [items, setItems] = useState<Notification[]>([]);

  const timeAgo = (iso: string) => {
    const diff = Date.now() - new Date(iso).getTime();
    const days = Math.floor(diff / 86_400_000);
    if (days <= 0) return t("notif.today");
    if (days === 1) return t("notif.yesterday");
    return t("notif.daysAgo", { n: days });
  };

  useEffect(() => {
    if (query.data) setItems(query.data);
  }, [query.data]);

  const unread = items.filter((n) => !n.read).length;
  const markAllRead = () => setItems((prev) => prev.map((n) => ({ ...n, read: true })));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-extrabold">
            {t("notif.title")}
            {unread > 0 && (
              <span className="grid size-6 place-items-center rounded-full bg-primary text-xs text-primary-foreground">
                {unread}
              </span>
            )}
          </h1>
          <p className="text-muted-foreground">{t("notif.subtitle")}</p>
        </div>
        {unread > 0 && (
          <Button variant="outline" size="sm" onClick={markAllRead}>
            <CheckCheck className="size-4" /> {t("notif.markAllRead")}
          </Button>
        )}
      </div>

      {query.isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : query.isError ? (
        <ErrorState onRetry={() => query.refetch()} />
      ) : items.length === 0 ? (
        <EmptyState title={t("notif.emptyTitle")} description={t("notif.emptyDesc")} />
      ) : (
        <Card className="divide-y overflow-hidden p-0">
          {items.map((n) => {
            const meta = ICONS[n.type];
            return (
              <button
                key={n.id}
                onClick={() => setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: true } : x)))}
                className={cn(
                  "flex w-full items-start gap-4 px-5 py-4 text-left transition-colors hover:bg-secondary/40",
                  !n.read && "bg-accent/30",
                )}
              >
                <div className={`grid size-10 shrink-0 place-items-center rounded-full ${meta.color}`}>
                  <meta.icon className="size-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="font-semibold">{n.title}</p>
                    {!n.read && <span className="size-2 rounded-full bg-primary" />}
                  </div>
                  <p className="mt-0.5 text-sm text-muted-foreground">{n.body}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{timeAgo(n.createdAt)}</p>
                </div>
              </button>
            );
          })}
        </Card>
      )}

      {!query.isLoading && items.length > 0 && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Bell className="size-3.5" /> {t("notif.prefHint")}
        </div>
      )}
    </div>
  );
}
