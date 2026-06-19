import { CourseForm } from "@/features/courses/course-form";

export default async function EditCoursePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-extrabold">Edit course</h1>
        <p className="text-muted-foreground">Course #{id}</p>
      </div>
      <CourseForm />
    </div>
  );
}
