"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  Briefcase,
  Camera,
  Code2,
  LineChart,
  Megaphone,
  Palette,
  Shapes,
  type LucideIcon,
} from "lucide-react";
import { Card } from "@/components/ui/card";
import { PopularCourses } from "@/features/courses/popular-courses";
import { LoadingState } from "@/components/shared/states";
import { categoriesService } from "@/services/categories.service";
import { ROUTES } from "@/constants";
import { formatNumber } from "@/lib/utils";
import { useLocale, useT } from "@/providers/locale-provider";
import type { Category } from "@/types";

// Ikonka/rang — sof UI bezagi, slug bo'yicha tanlanadi (backendda saqlanmaydi).
const STYLES: Record<string, { icon: LucideIcon; color: string }> = {
  development: { icon: Code2, color: "bg-indigo-100 text-indigo-600" },
  design: { icon: Palette, color: "bg-pink-100 text-pink-600" },
  business: { icon: Briefcase, color: "bg-emerald-100 text-emerald-600" },
  marketing: { icon: Megaphone, color: "bg-amber-100 text-amber-600" },
  "data-science": { icon: LineChart, color: "bg-sky-100 text-sky-600" },
  photography: { icon: Camera, color: "bg-violet-100 text-violet-600" },
};
const FALLBACK_STYLE = { icon: Shapes, color: "bg-slate-100 text-slate-600" };

function categoryName(c: Category, locale: string): string {
  if (locale === "uz") return c.nameUz;
  if (locale === "ru") return c.nameRu;
  return c.nameEn;
}

export default function CategoriesPage() {
  const t = useT();
  const { locale } = useLocale();
  const { data: categories, isLoading } = useQuery({
    queryKey: ["categories"],
    queryFn: categoriesService.list,
  });

  const all = categories ?? [];
  const parents = all.filter((c) => c.parentId == null);

  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="max-w-2xl">
        <h1 className="text-3xl font-extrabold">{t("categories.title")}</h1>
        <p className="mt-2 text-muted-foreground">{t("categories.subtitle")}</p>
      </div>

      {isLoading ? (
        <LoadingState className="min-h-[30vh]" />
      ) : (
        <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {parents.map((cat) => {
            const { icon: Icon, color } = STYLES[cat.slug] ?? FALLBACK_STYLE;
            const children = all.filter((c) => c.parentId === cat.id);
            const description = children.map((c) => categoryName(c, locale)).join(", ");
            return (
              <Link key={cat.slug} href={ROUTES.category(cat.slug)}>
                <Card className="h-full p-6 transition-shadow hover:shadow-md">
                  <div className={`grid size-12 place-items-center rounded-xl ${color}`}>
                    <Icon className="size-6" />
                  </div>
                  <h2 className="mt-4 text-lg font-bold">{categoryName(cat, locale)}</h2>
                  <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{description}</p>
                  <p className="mt-3 text-sm font-semibold text-primary">
                    {t("categories.coursesCount", { n: formatNumber(cat.courseCount ?? 0) })}
                  </p>
                </Card>
              </Link>
            );
          })}
        </div>
      )}

      <section className="mt-16">
        <div className="flex items-end justify-between">
          <h2 className="text-2xl font-extrabold">{t("categories.trending")}</h2>
          <Link href={ROUTES.courses} className="text-sm font-semibold text-primary hover:underline">
            {t("common.viewAll")}
          </Link>
        </div>
        <div className="mt-8">
          <PopularCourses />
        </div>
      </section>
    </div>
  );
}
