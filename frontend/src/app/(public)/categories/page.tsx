"use client";

import Link from "next/link";
import {
  Briefcase,
  Camera,
  Code2,
  LineChart,
  Megaphone,
  Palette,
  type LucideIcon,
} from "lucide-react";
import { Card } from "@/components/ui/card";
import { PopularCourses } from "@/features/courses/popular-courses";
import { categoryTree } from "@/services/mock/data";
import { ROUTES } from "@/constants";
import { formatNumber } from "@/lib/utils";
import { useT } from "@/providers/locale-provider";

const ICONS: Record<string, LucideIcon> = {
  Code2,
  Palette,
  Briefcase,
  Megaphone,
  LineChart,
  Camera,
};

export default function CategoriesPage() {
  const t = useT();
  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="max-w-2xl">
        <h1 className="text-3xl font-extrabold">{t("categories.title")}</h1>
        <p className="mt-2 text-muted-foreground">{t("categories.subtitle")}</p>
      </div>

      <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {categoryTree.map((cat) => {
          const Icon = ICONS[cat.icon] ?? Code2;
          const base = `cat.${cat.slug.replace(/\s/g, "")}`;
          return (
            <Link key={cat.slug} href={ROUTES.category(cat.slug)}>
              <Card className="h-full p-6 transition-shadow hover:shadow-md">
                <div className={`grid size-12 place-items-center rounded-xl ${cat.color}`}>
                  <Icon className="size-6" />
                </div>
                <h2 className="mt-4 text-lg font-bold">{t(base)}</h2>
                <p className="mt-1 text-sm text-muted-foreground">{t(`${base}.desc`)}</p>
                <p className="mt-3 text-sm font-semibold text-primary">
                  {t("categories.coursesCount", { n: formatNumber(cat.courseCount) })}
                </p>
              </Card>
            </Link>
          );
        })}
      </div>

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
