"use client";

import { useQuery } from "@tanstack/react-query";
import { categoriesService } from "@/services/categories.service";
import { useLocale } from "@/providers/locale-provider";
import type { Category } from "@/types";
import type { Locale } from "@/i18n/config";

/** Fallback: "web-development" -> "Web Development" (kategoriya hali yuklanmagan bo'lsa). */
function humanize(slug: string): string {
  return slug
    .split("-")
    .map((w) => (w ? w.charAt(0).toUpperCase() + w.slice(1) : w))
    .join(" ");
}

export function localizedCategoryName(category: Category, locale: Locale): string {
  if (locale === "uz") return category.nameUz || category.nameEn;
  if (locale === "ru") return category.nameRu || category.nameEn;
  return category.nameEn;
}

/**
 * Backend kurslarda kategoriya slug'ini qaytaradi; nomlar dinamik (admin
 * yaratadi), shuning uchun i18n kaliti emas — GET /categories dan olinadi.
 */
export function useCategoryName(): (slug: string) => string {
  const { locale } = useLocale();
  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: categoriesService.list,
    staleTime: 5 * 60_000,
  });

  return (slug: string) => {
    if (!slug) return "";
    const category = categories?.find((c) => c.slug === slug);
    return category ? localizedCategoryName(category, locale) : humanize(slug);
  };
}
