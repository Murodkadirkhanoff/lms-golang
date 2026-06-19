"use client";

import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CourseCard } from "./course-card";
import { CardGridSkeleton, EmptyState, ErrorState } from "@/components/shared/states";
import { Pagination } from "@/components/shared/pagination";
import { coursesService } from "@/services/courses.service";
import { CATEGORIES, SORT_OPTIONS } from "@/constants";
import type { CourseQuery } from "@/types";

const PAGE_SIZE = 8;

export function CourseCatalog() {
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("All");
  const [sort, setSort] = useState<CourseQuery["sort"]>("popular");
  const [page, setPage] = useState(1);

  const query: CourseQuery = { search, category, sort, page, pageSize: PAGE_SIZE };

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
        <h1 className="text-3xl font-extrabold">Explore courses</h1>
        <p className="mt-2 text-muted-foreground">{data ? `${data.total} courses` : "Browse our catalog"}</p>
      </div>

      {/* Filters */}
      <div className="mb-8 flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-2.5 size-5 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => resetPage(setSearch)(e.target.value)}
            placeholder="Search courses…"
            className="pl-10"
          />
        </div>
        <Select value={category} onValueChange={resetPage(setCategory)}>
          <SelectTrigger className="sm:w-48">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="All">All categories</SelectItem>
            {CATEGORIES.map((c) => (
              <SelectItem key={c} value={c}>
                {c}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={sort} onValueChange={(v) => resetPage(setSort)(v as CourseQuery["sort"])}>
          <SelectTrigger className="sm:w-52">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            {SORT_OPTIONS.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                {o.label}
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
        <EmptyState title="No courses found" description="Try adjusting your search or filters." />
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
