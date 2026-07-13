import type { Metadata } from "next";
import { CourseCatalog } from "@/features/courses/course-catalog";
import { CATEGORIES, SORT_OPTIONS } from "@/constants";
import type { CourseQuery } from "@/types";

export const metadata: Metadata = { title: "Courses" };

export default async function CoursesPage({
  searchParams,
}: {
  searchParams: Promise<{ search?: string; category?: string; sort?: string; page?: string }>;
}) {
  const { search, category, sort, page } = await searchParams;
  const validCategory = category && (CATEGORIES as readonly string[]).includes(category) ? category : "All";
  const validSort = SORT_OPTIONS.some((o) => o.value === sort)
    ? (sort as CourseQuery["sort"])
    : "popular";
  const parsedPage = Number(page);
  const validPage = Number.isInteger(parsedPage) && parsedPage > 0 ? parsedPage : 1;

  return (
    <CourseCatalog
      initialSearch={search ?? ""}
      initialCategory={validCategory}
      initialSort={validSort}
      initialPage={validPage}
    />
  );
}
