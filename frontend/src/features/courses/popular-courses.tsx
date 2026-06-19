"use client";

import { useQuery } from "@tanstack/react-query";
import { coursesService } from "@/services/courses.service";
import { CourseCard } from "./course-card";
import { CardGridSkeleton, ErrorState } from "@/components/shared/states";

export function PopularCourses() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["courses", "popular"],
    queryFn: () => coursesService.getPopular(4),
  });

  if (isLoading) return <CardGridSkeleton count={4} />;
  if (isError || !data) return <ErrorState onRetry={() => refetch()} />;

  return (
    <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
      {data.map((course) => (
        <CourseCard key={course.id} course={course} />
      ))}
    </div>
  );
}
