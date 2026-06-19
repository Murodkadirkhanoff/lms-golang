import type { Metadata } from "next";
import { CourseCatalog } from "@/features/courses/course-catalog";

export const metadata: Metadata = { title: "Courses" };

export default function CoursesPage() {
  return <CourseCatalog />;
}
