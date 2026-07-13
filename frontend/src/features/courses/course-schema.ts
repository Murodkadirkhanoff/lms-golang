import { z } from "zod";

// --- lessons table ---
// title, type, content_url | content, duration_seconds, position (from order),
// price, is_free. CHECK (is_free = false OR price = 0). Duration is collected in
// minutes in the UI and converted to duration_seconds when submitting.
// A lesson is either a "video" (uploaded file at contentUrl) or a "text"
// markdown article (content).
export const lessonSchema = z
  .object({
    title: z.string().min(1, "Lesson title is required").max(200, "Max 200 characters"),
    type: z.enum(["video", "text"]),
    contentUrl: z.string().optional(),
    content: z.string().optional(),
    durationMinutes: z.number().min(0, "Must be 0 or more"),
    price: z.number().min(0, "Price cannot be negative"),
    isFree: z.boolean(),
  })
  .refine((l) => !l.isFree || l.price === 0, {
    message: "Free lessons must be priced 0",
    path: ["price"],
  })
  .refine((l) => l.type !== "video" || (l.contentUrl?.trim().length ?? 0) > 0, {
    message: "Upload a video for this lesson",
    path: ["contentUrl"],
  })
  .refine((l) => l.type !== "text" || (l.content?.trim().length ?? 0) > 0, {
    message: "Text lessons need content",
    path: ["content"],
  });

// --- modules table ---
// title, position (from order). Belongs to a course.
export const moduleSchema = z.object({
  title: z.string().min(1, "Section title is required").max(200, "Max 200 characters"),
  lessons: z.array(lessonSchema),
});

// --- courses table ---
// instructor_id is set server-side from the authenticated user; slug is
// generated server-side from the title (UNIQUE).
export const courseSchema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters").max(200, "Max 200 characters"),
  description: z.string().max(5000, "Max 5000 characters").optional(),
  categoryId: z.number().int().positive().nullable(),
  lang: z.enum(["uz", "ru", "en"]),
  price: z.number().min(0, "Price cannot be negative"),
  isPublished: z.boolean(),
  modules: z.array(moduleSchema),
});

export type LessonFormValues = z.infer<typeof lessonSchema>;
export type ModuleFormValues = z.infer<typeof moduleSchema>;
export type CourseFormValues = z.infer<typeof courseSchema>;
