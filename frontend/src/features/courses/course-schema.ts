import { z } from "zod";

export const courseSchema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters").max(200, "Max 200 characters"),
  description: z.string().min(10, "Add a longer description").max(5000, "Max 5000 characters"),
  category: z.string().min(1, "Select a category"),
  lang: z.enum(["uz", "ru", "en"]),
  price: z.number().min(0, "Price cannot be negative"),
  isPublished: z.boolean(),
});

export type CourseFormValues = z.infer<typeof courseSchema>;
