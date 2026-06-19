export const APP_NAME = "LearnHub";

export const ROUTES = {
  home: "/",
  courses: "/courses",
  course: (slug: string) => `/courses/${slug}`,
  login: "/login",
  register: "/register",
  forgotPassword: "/forgot-password",
  // Learner area
  dashboard: "/dashboard",
  learn: (id: number | string) => `/learn/${id}`,
  quiz: (id: number | string) => `/quiz/${id}`,
  certificates: "/certificates",
  profile: "/profile",
  // Studio (any user can teach — no role gate)
  studio: "/studio",
  studioCourses: "/studio/courses",
  studioCourseNew: "/studio/courses/new",
  studioCourseEdit: (id: number | string) => `/studio/courses/${id}`,
  studioAnalytics: "/studio/analytics",
  // Admin
  admin: "/admin",
  adminUsers: "/admin/users",
  adminCourses: "/admin/courses",
  adminCategories: "/admin/categories",
} as const;

export const CATEGORIES = [
  "Development",
  "Design",
  "Business",
  "Marketing",
  "Data Science",
  "Photography",
] as const;

export const LANGUAGES = [
  { value: "uz", label: "Uzbek" },
  { value: "ru", label: "Russian" },
  { value: "en", label: "English" },
] as const;

export const SORT_OPTIONS = [
  { value: "popular", label: "Most popular" },
  { value: "newest", label: "Newest" },
  { value: "price-asc", label: "Price: low to high" },
  { value: "price-desc", label: "Price: high to low" },
] as const;

export const THUMB_COLORS = [
  "from-indigo-400 to-violet-500",
  "from-emerald-400 to-teal-500",
  "from-amber-400 to-orange-500",
  "from-rose-400 to-pink-500",
  "from-sky-400 to-blue-500",
  "from-fuchsia-400 to-purple-500",
];
