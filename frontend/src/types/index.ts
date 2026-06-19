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

export interface Lesson {
  id: number;
  title: string;
  durationSeconds: number;
  isFree: boolean;
  contentUrl?: string;
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
