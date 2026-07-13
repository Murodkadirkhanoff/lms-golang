"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { Heart, ShoppingCart, Star, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LoadingState } from "@/components/shared/states";
import { coursesService } from "@/services/courses.service";
import { useWishlist } from "@/features/wishlist/wishlist-context";
import { useCart } from "@/features/cart/cart-context";
import { useT } from "@/providers/locale-provider";
import { ROUTES } from "@/constants";
import { formatNumber, formatPrice } from "@/lib/utils";

export default function WishlistPage() {
  const wishlist = useWishlist();
  const cart = useCart();
  const t = useT();
  const { data: courses, isLoading } = useQuery({
    queryKey: ["wishlist", wishlist.ids],
    queryFn: () => coursesService.getByIds(wishlist.ids),
    enabled: wishlist.ids.length > 0,
  });

  const items = courses ?? [];

  return (
    <div className="mx-auto max-w-7xl px-6 py-10">
      <div className="flex items-center gap-3">
        <Heart className="size-7 text-rose-500" />
        <h1 className="text-3xl font-extrabold">{t("wishlist.title")}</h1>
      </div>
      <p className="mt-1 text-muted-foreground">{t("wishlist.count", { n: wishlist.count })}</p>

      {wishlist.ids.length === 0 ? (
        <Card className="mt-8 flex flex-col items-center gap-4 p-16 text-center">
          <div className="grid size-14 place-items-center rounded-full bg-rose-100 text-rose-500">
            <Heart className="size-6" />
          </div>
          <div>
            <p className="text-lg font-bold">{t("wishlist.emptyTitle")}</p>
            <p className="mt-1 text-sm text-muted-foreground">{t("wishlist.emptyDesc")}</p>
          </div>
          <Button asChild>
            <Link href={ROUTES.courses}>{t("home.browseCourses")}</Link>
          </Button>
        </Card>
      ) : isLoading ? (
        <LoadingState className="min-h-[40vh]" />
      ) : (
        <div className="mt-8 space-y-4">
          {items.map((course) => (
            <Card key={course.id} className="flex flex-col gap-4 p-4 sm:flex-row sm:items-center">
              <Link
                href={ROUTES.course(course.slug)}
                className={`h-28 w-full shrink-0 rounded-lg bg-gradient-to-br sm:h-20 sm:w-32 ${course.thumbnailColor}`}
              />
              <div className="min-w-0 flex-1">
                <Link href={ROUTES.course(course.slug)} className="font-bold hover:text-primary">
                  {course.title}
                </Link>
                <p className="text-xs text-muted-foreground">{t("common.by")} {course.instructor.name}</p>
                <div className="mt-1 flex items-center gap-1 text-xs">
                  <Star className="size-3.5 fill-amber-400 text-amber-400" />
                  <span className="font-semibold">{course.rating}</span>
                  <span className="text-muted-foreground">({formatNumber(course.ratingCount)})</span>
                  <Badge variant="secondary" className="ml-2">{t(`cat.${course.category.replace(/\s/g, "")}`)}</Badge>
                </div>
              </div>
              <div className="flex items-center gap-3 sm:flex-col sm:items-end">
                <span className="text-lg font-extrabold">{formatPrice(course.price)}</span>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    onClick={() => {
                      cart.addCourse(course.id);
                      wishlist.remove(course.id);
                    }}
                  >
                    <ShoppingCart className="size-4" /> {t("detail.addToCart")}
                  </Button>
                  <Button size="icon" variant="ghost" aria-label={t("common.remove")} onClick={() => wishlist.remove(course.id)}>
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
