export type Lang = "uz" | "ru" | "en";

export interface User {
  id: number;
  name: string;
  email: string;
  avatarColor?: string;
  // No role gating between learning and teaching — any user can do both.
  // "admin" additionally unlocks the admin panel link.
  role?: "student" | "instructor" | "admin";
  createdAt: string;
}

export interface Category {
  id: number;
  slug: string;
  nameEn: string;
  nameUz: string;
  nameRu: string;
  parentId?: number | null;
  // Published kurslar soni (ota kategoriyada bolalarniki bilan birga).
  courseCount?: number;
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
  // Paywall: kontent yashirilgan — sotib olinmagan pullik dars.
  locked?: boolean;
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
  /** Yuklangan rasm URL'i; bo'sh bo'lsa thumbnailColor gradienti ko'rsatiladi. */
  thumbnailUrl?: string;
  categoryId?: number | null;
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
  // Progress yozish uchun (PATCH /enrollments/{id}/progress). Mock rejimda 0.
  enrollmentId: number;
  course: Course;
  progress: number; // 0-100
  currentLesson: string;
  lessonsCompleted: number;
  completedLessonIds?: number[];
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

export interface QuizAttempt {
  id: number;
  createdAt: string;
  score: number;
}

// GET /me/teaching/stats javobi — studio dashboard va analytics.
export interface TeachingStats {
  totalRevenue: number;
  monthlyRevenue: { month: string; revenue: number }[]; // month: "YYYY-MM"
  totalStudents: number;
  activeStudents: number;
  publishedCourses: number;
  draftCourses: number;
  avgRating: number;
  avgCompletion: number; // 0-100
  avgQuizScore: number; // 0-100
  engagement: { courseId: number; title: string; students: number; completion: number }[];
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
