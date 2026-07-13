"use client";

import Link from "next/link";
import { Heart, Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ROUTES } from "@/constants";
import { cn, formatNumber, formatPrice } from "@/lib/utils";
import { useWishlist } from "@/features/wishlist/wishlist-context";
import { useT } from "@/providers/locale-provider";
import { useToast } from "@/providers/toast-provider";
import type { Course } from "@/types";

export function CourseCard({ course }: { course: Course }) {
  const wishlist = useWishlist();
  const t = useT();
  const toast = useToast();
  const saved = wishlist.has(course.id);

  return (
    <Link href={ROUTES.course(course.slug)} className="group">
      <Card className="relative overflow-hidden transition-shadow hover:shadow-lg">
        <button
          type="button"
          aria-label={saved ? "Remove from wishlist" : "Add to wishlist"}
          onClick={(e) => {
            e.preventDefault();
            wishlist.toggle(course.id);
            toast.success(saved ? t("toast.wishlistRemoved") : t("toast.wishlistAdded"));
          }}
          className="absolute right-3 top-3 z-10 grid size-9 place-items-center rounded-full bg-background/90 text-foreground shadow-sm backdrop-blur transition-colors hover:bg-background"
        >
          <Heart className={cn("size-4", saved && "fill-rose-500 text-rose-500")} />
        </button>
        <div className={`aspect-video bg-gradient-to-br ${course.thumbnailColor}`} />
        <div className="space-y-2 p-5">
          <Badge>{t(`cat.${course.category.replace(/\s/g, "")}`)}</Badge>
          <h3 className="line-clamp-2 font-bold leading-snug group-hover:text-primary">{course.title}</h3>
          <p className="text-sm text-muted-foreground">{t("common.by")} {course.instructor.name}</p>
          <div className="flex items-center gap-1 text-sm">
            <Star className="size-4 fill-amber-400 text-amber-400" />
            <span className="font-semibold">{course.rating}</span>
            <span className="text-muted-foreground">({formatNumber(course.ratingCount)})</span>
          </div>
          <div className="pt-1 text-lg font-extrabold">{formatPrice(course.price)}</div>
        </div>
      </Card>
    </Link>
  );
}
