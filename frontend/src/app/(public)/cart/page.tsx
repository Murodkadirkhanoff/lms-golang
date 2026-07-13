"use client";

import Link from "next/link";
import { BookOpen, PlayCircle, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingState } from "@/components/shared/states";
import { useCart } from "@/features/cart/cart-context";
import { useCartLines } from "@/features/cart/use-cart-lines";
import { useT } from "@/providers/locale-provider";
import { ROUTES } from "@/constants";
import { formatPrice } from "@/lib/utils";

export default function CartPage() {
  const cart = useCart();
  const t = useT();
  const { lines, subtotal, isLoading } = useCartLines();

  const originalTotal = lines.reduce((sum, l) => sum + (l.price > 0 ? l.price * 1.6 : 0), 0);
  const savings = originalTotal - subtotal;

  return (
    <div className="mx-auto max-w-7xl px-6 py-10">
      <h1 className="text-3xl font-extrabold">{t("cart.title")}</h1>
      <p className="mt-1 text-muted-foreground">{t("cart.count", { n: cart.count })}</p>

      {cart.count === 0 ? (
        <Card className="mt-8 flex flex-col items-center gap-4 p-16 text-center">
          <div className="grid size-14 place-items-center rounded-full bg-secondary text-muted-foreground">🛒</div>
          <div>
            <p className="text-lg font-bold">{t("cart.emptyTitle")}</p>
            <p className="mt-1 text-sm text-muted-foreground">{t("cart.emptyDesc")}</p>
          </div>
          <Button asChild>
            <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
          </Button>
        </Card>
      ) : isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : (
        <div className="mt-8 grid gap-8 lg:grid-cols-3">
          <div className="space-y-4 lg:col-span-2">
            {lines.map((line) => (
              <Card key={line.key} className="flex gap-4 p-4">
                <Link
                  href={ROUTES.course(line.slug)}
                  className={`grid h-24 w-40 shrink-0 place-items-center rounded-lg bg-gradient-to-br ${line.thumbnailColor} text-white`}
                >
                  {line.kind === "course" ? <BookOpen className="size-7" /> : <PlayCircle className="size-7" />}
                </Link>
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <Link href={ROUTES.course(line.slug)} className="font-bold hover:text-primary">
                        {line.title}
                      </Link>
                      <p className="text-xs text-muted-foreground">
                        {line.kind === "course"
                          ? `${t("common.by")} ${line.subtitle}`
                          : line.subtitle}
                      </p>
                      <Badge variant={line.kind === "course" ? "secondary" : "outline"} className="mt-1">
                        {line.kind === "course" ? t("cart.fullCourse") : t("cart.singleLesson")}
                      </Badge>
                    </div>
                    <div className="text-right text-lg font-extrabold">{formatPrice(line.price)}</div>
                  </div>
                  <div className="mt-3 flex gap-4 text-xs font-medium">
                    <button
                      onClick={() =>
                        line.kind === "course"
                          ? cart.removeCourse(line.courseId)
                          : cart.removeLesson(line.lessonId!)
                      }
                      className="flex items-center gap-1 text-rose-600 hover:underline"
                    >
                      <Trash2 className="size-3.5" /> {t("common.remove")}
                    </button>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          {/* Summary */}
          <aside>
            <div className="lg:sticky lg:top-24">
              <Card className="p-6">
                <p className="text-sm text-muted-foreground">{t("cart.total")}</p>
                <p className="text-3xl font-extrabold">{formatPrice(subtotal)}</p>
                {savings > 0 && (
                  <p className="mt-1 text-sm text-muted-foreground">
                    <span className="line-through">{formatPrice(originalTotal)}</span>{" "}
                    <span className="font-semibold text-emerald-600">{t("cart.save", { x: formatPrice(savings) })}</span>
                  </p>
                )}
                <Button asChild className="mt-4 w-full" size="lg">
                  <Link href={ROUTES.checkout}>{t("cart.checkout")}</Link>
                </Button>

                <div className="mt-6">
                  <p className="text-sm font-semibold">{t("cart.promotions")}</p>
                  <div className="mt-2 flex gap-2">
                    <Input placeholder={t("cart.coupon")} className="h-10" />
                    <Button variant="outline">{t("common.apply")}</Button>
                  </div>
                </div>
              </Card>
            </div>
          </aside>
        </div>
      )}
    </div>
  );
}
