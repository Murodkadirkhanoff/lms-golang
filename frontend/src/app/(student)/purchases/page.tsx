"use client";

import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Download } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { LoadingState, ErrorState, EmptyState } from "@/components/shared/states";
import { Pagination } from "@/components/shared/pagination";
import { ordersService } from "@/services/orders.service";
import { formatDate, formatPrice } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";
import type { OrderStatus } from "@/types";

const STATUS_VARIANT: Record<OrderStatus, "success" | "secondary" | "warning"> = {
  completed: "success",
  refunded: "secondary",
  pending: "warning",
};

export default function PurchasesPage() {
  const t = useT();
  const [page, setPage] = useState(1);
  const orders = useQuery({
    queryKey: ["orders", page],
    queryFn: () => ordersService.list(page),
    placeholderData: keepPreviousData,
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">{t("purchases.title")}</h1>
        <p className="text-muted-foreground">{t("purchases.subtitle")}</p>
      </div>

      {orders.isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : orders.isError ? (
        <ErrorState onRetry={() => orders.refetch()} />
      ) : !orders.data || orders.data.items.length === 0 ? (
        <EmptyState title={t("purchases.emptyTitle")} description={t("purchases.emptyDesc")} />
      ) : (
        <div className="space-y-4">
          {orders.data.items.map((order) => (
            <Card key={order.id} className="overflow-hidden">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b bg-secondary/40 px-5 py-3">
                <div className="flex flex-wrap items-center gap-x-6 gap-y-1 text-sm">
                  <div>
                    <span className="text-muted-foreground">{t("purchases.order")} </span>
                    <span className="font-semibold">{order.id}</span>
                  </div>
                  <div className="text-muted-foreground">{formatDate(order.date)}</div>
                  <div className="text-muted-foreground">{order.paymentMethod}</div>
                  <Badge variant={STATUS_VARIANT[order.status]}>
                    {t(`status.${order.status}`)}
                  </Badge>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-lg font-extrabold">{formatPrice(order.total)}</span>
                  <Button variant="outline" size="sm">
                    <Download className="size-4" /> {t("purchases.receipt")}
                  </Button>
                </div>
              </div>
              <ul className="divide-y">
                {order.items.map((item, idx) => (
                  <li key={idx} className="flex items-center gap-4 px-5 py-4">
                    <div className={`h-12 w-20 shrink-0 rounded-md bg-gradient-to-br ${item.thumbnailColor}`} />
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-semibold">{item.courseTitle}</p>
                      <p className="text-xs text-muted-foreground">{t("common.by")} {item.instructor}</p>
                    </div>
                    <span className="text-sm font-semibold">{formatPrice(item.price)}</span>
                  </li>
                ))}
              </ul>
            </Card>
          ))}
          <Pagination
            page={orders.data.page}
            pageSize={orders.data.pageSize}
            total={orders.data.total}
            onPageChange={setPage}
          />
        </div>
      )}
    </div>
  );
}
