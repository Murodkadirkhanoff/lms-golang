"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CourseCard } from "./course-card";
import { CardGridSkeleton, EmptyState, ErrorState } from "@/components/shared/states";
import { Pagination } from "@/components/shared/pagination";
import { coursesService } from "@/services/courses.service";
import { categoriesService } from "@/services/categories.service";
import { SORT_OPTIONS } from "@/constants";
import { useLocale } from "@/providers/locale-provider";
import { localizedCategoryName } from "./use-category-name";
import type { CourseQuery } from "@/types";

const PAGE_SIZE = 8;

interface CourseCatalogProps {
  initialSearch?: string;
  initialCategory?: string;
  initialSort?: CourseQuery["sort"];
  initialPage?: number;
}

export function CourseCatalog({
  initialSearch = "",
  initialCategory = "All",
  initialSort = "popular",
  initialPage = 1,
}: CourseCatalogProps) {
  const { locale, t } = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const [search, setSearch] = useState(initialSearch);
  // Debounced copy of `search` — only this drives the query, so typing doesn't
  // fire a request (or a URL update) on every keystroke.
  const [debouncedSearch, setDebouncedSearch] = useState(initialSearch);
  const [category, setCategory] = useState(initialCategory);
  const [sort, setSort] = useState<CourseQuery["sort"]>(initialSort);
  const [page, setPage] = useState(initialPage);

  useEffect(() => {
    const id = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(id);
  }, [search]);

  // Reflect the active filters in the URL so results are shareable and the
  // browser back button restores previous filter states.
  useEffect(() => {
    const params = new URLSearchParams();
    if (debouncedSearch) params.set("search", debouncedSearch);
    if (category !== "All") params.set("category", category);
    if (sort && sort !== "popular") params.set("sort", sort);
    if (page > 1) params.set("page", String(page));
    const qs = params.toString();
    router.replace(qs ? `${pathname}?${qs}` : pathname, { scroll: false });
  }, [debouncedSearch, category, sort, page, pathname, router]);

  const query: CourseQuery = { search: debouncedSearch, category, sort, page, pageSize: PAGE_SIZE };

  // Filtr real kategoriyalardan quriladi (qiymat — backend kutadigan slug).
  const { data: allCategories } = useQuery({
    queryKey: ["categories"],
    queryFn: categoriesService.list,
    staleTime: 5 * 60_000,
  });
  const parentCategories = (allCategories ?? []).filter((c) => c.parentId == null);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["courses", query],
    queryFn: () => coursesService.list(query),
    placeholderData: keepPreviousData,
  });

  const resetPage = <T,>(setter: (v: T) => void) => (v: T) => {
    setter(v);
    setPage(1);
  };

  return (
    <div className="mx-auto max-w-7xl px-6 py-10">
      <div className="mb-8">
        <h1 className="text-3xl font-extrabold">{t("catalog.title")}</h1>
        <p className="mt-2 text-muted-foreground">
          {data ? t("catalog.count", { n: data.total }) : t("catalog.browseHint")}
        </p>
      </div>

      {/* Filters */}
      <div className="mb-8 flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-2.5 size-5 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => resetPage(setSearch)(e.target.value)}
            placeholder={t("catalog.searchPlaceholder")}
            className="pl-10"
          />
        </div>
        <Select value={category} onValueChange={resetPage(setCategory)}>
          <SelectTrigger className="sm:w-48">
            <SelectValue placeholder={t("catalog.category")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="All">{t("catalog.allCategories")}</SelectItem>
            {parentCategories.map((c) => (
              <SelectItem key={c.slug} value={c.slug}>
                {localizedCategoryName(c, locale)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={sort} onValueChange={(v) => resetPage(setSort)(v as CourseQuery["sort"])}>
          <SelectTrigger className="sm:w-52">
            <SelectValue placeholder={t("catalog.sortBy")} />
          </SelectTrigger>
          <SelectContent>
            {SORT_OPTIONS.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                {t(`sort.${o.value}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Results */}
      {isLoading ? (
        <CardGridSkeleton count={8} />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.items.length === 0 ? (
        <EmptyState title={t("states.noCoursesTitle")} description={t("states.noCoursesDesc")} />
      ) : (
        <>
          <div className={`grid gap-6 sm:grid-cols-2 lg:grid-cols-4 ${isFetching ? "opacity-60" : ""}`}>
            {data.items.map((course) => (
              <CourseCard key={course.id} course={course} />
            ))}
          </div>
          <div className="mt-10">
            <Pagination page={data.page} pageSize={data.pageSize} total={data.total} onPageChange={setPage} />
          </div>
        </>
      )}
    </div>
  );
}
