"use client";

import { useQuery } from "@tanstack/react-query";
import { coursesService } from "@/services/courses.service";
import type { Course } from "@/types";
import { useCart, entryKey, type CartEntry } from "./cart-context";

// A resolved, renderable cart line: course or lesson flattened to display data.
export interface CartLine {
  key: string;
  kind: "course" | "lesson";
  courseId: number;
  lessonId?: number;
  title: string;
  subtitle: string;
  price: number;
  thumbnailColor: string;
  slug: string;
}

function toLine(entry: CartEntry, course: Course): CartLine | null {
  if (entry.kind === "course") {
    return {
      key: entryKey(entry),
      kind: "course",
      courseId: course.id,
      title: course.title,
      subtitle: course.instructor.name,
      price: course.price,
      thumbnailColor: course.thumbnailColor,
      slug: course.slug,
    };
  }
  const lesson = (course.modules ?? []).flatMap((m) => m.lessons).find((l) => l.id === entry.lessonId);
  if (!lesson) return null;
  return {
    key: entryKey(entry),
    kind: "lesson",
    courseId: course.id,
    lessonId: lesson.id,
    title: lesson.title,
    subtitle: course.title,
    price: lesson.price,
    thumbnailColor: course.thumbnailColor,
    slug: course.slug,
  };
}

// Resolves cart entries into renderable lines by fetching each referenced
// course once. Shared by the cart and checkout pages so the math stays in sync.
export function useCartLines() {
  const cart = useCart();
  const courseIds = [...new Set(cart.items.map((i) => i.courseId))];

  const { data: courses, isLoading } = useQuery({
    queryKey: ["cart-courses", courseIds],
    queryFn: () => coursesService.getByIds(courseIds),
    enabled: courseIds.length > 0,
  });

  const byId = new Map((courses ?? []).map((c) => [c.id, c]));
  const lines = cart.items
    .map((entry) => {
      const course = byId.get(entry.courseId);
      return course ? toLine(entry, course) : null;
    })
    .filter((l): l is CartLine => l !== null);

  const subtotal = lines.reduce((sum, l) => sum + l.price, 0);

  return { lines, subtotal, isLoading: courseIds.length > 0 && isLoading };
}
