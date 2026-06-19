import { CourseForm } from "@/features/courses/course-form";

export default function NewCoursePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">Create a new course</h1>
        <p className="text-muted-foreground">Fill in the details — you can edit everything later.</p>
      </div>
      <CourseForm />
    </div>
  );
}
