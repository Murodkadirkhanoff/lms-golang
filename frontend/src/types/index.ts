export type Lang = "uz" | "ru" | "en";

export interface User {
  id: number;
  name: string;
  email: string;
  avatarColor?: string;
  // No role gating: any user can both learn and teach.
  createdAt: string;
}

export interface Category {
  id: number;
  slug: string;
  nameEn: string;
  nameUz: string;
  nameRu: string;
  parentId?: number | null;
}

export interface Instructor {
  id: number;
  name: string;
  headline: string;
  avatarColor: string;
  students: number;
  courses: number;
  rating: number;
}

export type LessonType = "video" | "text";

export interface Lesson {
  id: number;
  title: string;
  // "video": streamed/uploaded video at contentUrl.
  // "text": markdown article stored in `content`.
  type: LessonType;
  durationSeconds: number;
  isFree: boolean;
  // Per-lesson price. 0 when the lesson is free or only sold as part of the
  // course. Mirrors the `lessons.price` column; free lessons enforce price 0.
  price: number;
  contentUrl?: string; // video lessons
  content?: string; // text lessons (markdown)
  completed?: boolean;
}

export interface Module {
  id: number;
  title: string;
  lessons: Lesson[];
}

export interface Review {
  id: number;
  user: string;
  avatarColor: string;
  rating: number;
  comment: string;
  createdAt: string;
}

export interface Course {
  id: number;
  slug: string;
  title: string;
  description: string;
  thumbnailColor: string;
  category: string;
  lang: Lang;
  price: number;
  rating: number;
  ratingCount: number;
  studentCount: number;
  isPublished: boolean;
  instructor: Instructor;
  modules?: Module[];
  reviews?: Review[];
  totalLessons: number;
  totalDurationMinutes: number;
  createdAt: string;
}

export interface EnrolledCourse {
  course: Course;
  progress: number; // 0-100
  currentLesson: string;
  lessonsCompleted: number;
}

export interface Certificate {
  id: number;
  courseTitle: string;
  issuedAt: string;
  color: string;
}

export interface QuizQuestion {
  id: number;
  question: string;
  options: string[];
  correctIndex: number;
}

export interface Quiz {
  id: number;
  title: string;
  passingScore: number;
  timeLimitMinutes: number;
  questions: QuizQuestion[];
}

export interface Paginated<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}

export interface CourseQuery {
  search?: string;
  category?: string;
  sort?: "popular" | "newest" | "price-asc" | "price-desc";
  page?: number;
  pageSize?: number;
}

export interface CartItem {
  courseId: number;
  addedAt: string;
}

export type OrderStatus = "completed" | "refunded" | "pending";

export interface OrderItem {
  courseTitle: string;
  instructor: string;
  thumbnailColor: string;
  price: number;
}

export interface Order {
  id: string;
  date: string;
  status: OrderStatus;
  paymentMethod: string;
  items: OrderItem[];
  total: number;
}

export type NotificationType = "course" | "message" | "promo" | "system";

export interface Notification {
  id: number;
  type: NotificationType;
  title: string;
  body: string;
  createdAt: string;
  read: boolean;
}

export interface CategoryNode {
  slug: string;
  name: string;
  icon: string; // lucide icon name handled by the page
  color: string;
  courseCount: number;
  description: string;
}
