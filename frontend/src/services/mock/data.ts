import type {
  Certificate,
  Course,
  EnrolledCourse,
  Instructor,
  Quiz,
} from "@/types";

export const instructors: Instructor[] = [
  { id: 1, name: "Sarah Mitchell", headline: "Senior Software Engineer · Ex-Vercel", avatarColor: "bg-indigo-200", students: 48200, courses: 14, rating: 4.9 },
  { id: 2, name: "James Carter", headline: "Product Designer · Design Lead", avatarColor: "bg-emerald-200", students: 31400, courses: 9, rating: 4.8 },
  { id: 3, name: "Elena Rodriguez", headline: "Head of Product", avatarColor: "bg-amber-200", students: 18900, courses: 6, rating: 4.7 },
  { id: 4, name: "Michael Chen", headline: "Growth Marketer", avatarColor: "bg-rose-200", students: 52300, courses: 11, rating: 4.9 },
];

export const courses: Course[] = [
  {
    id: 1,
    slug: "complete-next-js-16-developer-course",
    title: "Complete Next.js 16 Developer Course",
    description:
      "Master the App Router, Server Components, Server Actions, caching and deployment by building real-world full-stack applications from scratch.",
    thumbnailColor: "from-indigo-400 to-violet-500",
    category: "Development",
    lang: "en",
    price: 499000,
    rating: 4.9,
    ratingCount: 2340,
    studentCount: 12840,
    isPublished: true,
    instructor: instructors[0],
    totalLessons: 42,
    totalDurationMinutes: 750,
    createdAt: "2026-06-01",
    modules: [
      {
        id: 1,
        title: "Getting Started",
        lessons: [
          { id: 1, title: "Introduction & setup", type: "video", durationSeconds: 500, isFree: true, price: 0, completed: true, contentUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4" },
          { id: 2, title: "Project structure", type: "video", durationSeconds: 725, isFree: false, price: 99000, completed: true, contentUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4" },
          {
            id: 3,
            title: "The App Router explained (reading)",
            type: "text",
            durationSeconds: 0,
            isFree: false,
            price: 129000,
            content: `## The App Router

The **App Router** is built on React Server Components and introduces a new
file-system based routing model under the \`app/\` directory.

### Key ideas

- Every folder maps to a **route segment**.
- \`layout.tsx\` wraps child routes and **persists** across navigations.
- \`page.tsx\` makes a segment publicly routable.

> Server Components render on the server by default — no JS shipped unless you opt in with \`"use client"\`.

\`\`\`tsx
export default function Page() {
  return <h1>Hello App Router</h1>;
}
\`\`\`

See the [official docs](https://nextjs.org) for more.`,
          },
        ],
      },
      {
        id: 2,
        title: "Routing & Layouts",
        lessons: [
          { id: 4, title: "Pages & routing basics", type: "video", durationSeconds: 610, isFree: false, price: 99000, completed: true, contentUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4" },
          { id: 5, title: "Nested routes & layouts", type: "video", durationSeconds: 880, isFree: false, price: 99000, contentUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4" },
          { id: 6, title: "Dynamic segments", type: "video", durationSeconds: 690, isFree: false, price: 129000, contentUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerFun.mp4" },
        ],
      },
    ],
    reviews: [
      { id: 1, user: "Laura Bennett", avatarColor: "bg-emerald-200", rating: 5, comment: "Best Next.js course out there. The Server Actions section alone was worth it.", createdAt: "2026-06-05" },
      { id: 2, user: "David Park", avatarColor: "bg-indigo-200", rating: 5, comment: "Finally understand caching in Next.js. Clear and practical.", createdAt: "2026-05-18" },
    ],
  },
  {
    id: 2,
    slug: "ui-ux-design-masterclass-2026",
    title: "UI/UX Design Masterclass 2026",
    description: "Learn modern product design: research, wireframing, design systems and prototyping in Figma.",
    thumbnailColor: "from-emerald-400 to-teal-500",
    category: "Design",
    lang: "en",
    price: 399000,
    rating: 4.8,
    ratingCount: 1820,
    studentCount: 9600,
    isPublished: true,
    instructor: instructors[1],
    totalLessons: 30,
    totalDurationMinutes: 540,
    createdAt: "2026-05-20",
  },
  {
    id: 3,
    slug: "data-driven-product-management",
    title: "Data-Driven Product Management",
    description: "Make better product decisions with metrics, experimentation and roadmapping.",
    thumbnailColor: "from-amber-400 to-orange-500",
    category: "Business",
    lang: "en",
    price: 599000,
    rating: 4.7,
    ratingCount: 980,
    studentCount: 5400,
    isPublished: true,
    instructor: instructors[2],
    totalLessons: 28,
    totalDurationMinutes: 600,
    createdAt: "2026-04-12",
  },
  {
    id: 4,
    slug: "digital-marketing-growth-hacking",
    title: "Digital Marketing & Growth Hacking",
    description: "Acquire and retain users with proven growth strategies across channels.",
    thumbnailColor: "from-rose-400 to-pink-500",
    category: "Marketing",
    lang: "en",
    price: 449000,
    rating: 4.9,
    ratingCount: 3100,
    studentCount: 18200,
    isPublished: true,
    instructor: instructors[3],
    totalLessons: 36,
    totalDurationMinutes: 680,
    createdAt: "2026-03-30",
  },
  {
    id: 5,
    slug: "python-for-data-science",
    title: "Python for Data Science",
    description: "From pandas to machine learning — a hands-on path into data science with Python.",
    thumbnailColor: "from-sky-400 to-blue-500",
    category: "Data Science",
    lang: "en",
    price: 0,
    rating: 4.6,
    ratingCount: 4200,
    studentCount: 32100,
    isPublished: true,
    instructor: instructors[0],
    totalLessons: 48,
    totalDurationMinutes: 900,
    createdAt: "2026-02-15",
  },
  {
    id: 6,
    slug: "advanced-typescript-deep-dive",
    title: "Advanced TypeScript Deep Dive",
    description: "Master generics, conditional types and type-level programming for production codebases.",
    thumbnailColor: "from-fuchsia-400 to-purple-500",
    category: "Development",
    lang: "en",
    price: 349000,
    rating: 4.8,
    ratingCount: 1440,
    studentCount: 8700,
    isPublished: true,
    instructor: instructors[0],
    totalLessons: 32,
    totalDurationMinutes: 520,
    createdAt: "2026-06-10",
  },
  {
    id: 7,
    slug: "react-native-mobile-apps",
    title: "Build Mobile Apps with React Native",
    description: "Ship iOS and Android apps from a single codebase — navigation, native modules, and app store publishing.",
    thumbnailColor: "from-cyan-400 to-sky-500",
    category: "Development",
    lang: "en",
    price: 549000,
    rating: 4.7,
    ratingCount: 1260,
    studentCount: 7400,
    isPublished: true,
    instructor: instructors[0],
    totalLessons: 38,
    totalDurationMinutes: 710,
    createdAt: "2026-05-02",
  },
  {
    id: 8,
    slug: "figma-design-systems",
    title: "Design Systems in Figma",
    description: "Build scalable component libraries, tokens and auto-layout patterns used by world-class product teams.",
    thumbnailColor: "from-pink-400 to-rose-500",
    category: "Design",
    lang: "en",
    price: 299000,
    rating: 4.9,
    ratingCount: 2010,
    studentCount: 11200,
    isPublished: true,
    instructor: instructors[1],
    totalLessons: 26,
    totalDurationMinutes: 430,
    createdAt: "2026-06-18",
  },
  {
    id: 9,
    slug: "startup-finance-fundraising",
    title: "Startup Finance & Fundraising",
    description: "Model your business, understand cap tables, and raise your first round with confidence.",
    thumbnailColor: "from-emerald-400 to-green-500",
    category: "Business",
    lang: "en",
    price: 649000,
    rating: 4.6,
    ratingCount: 720,
    studentCount: 3900,
    isPublished: true,
    instructor: instructors[2],
    totalLessons: 22,
    totalDurationMinutes: 480,
    createdAt: "2026-04-28",
  },
  {
    id: 10,
    slug: "seo-content-marketing-2026",
    title: "SEO & Content Marketing 2026",
    description: "Rank on Google, build topical authority and turn organic traffic into customers.",
    thumbnailColor: "from-orange-400 to-amber-500",
    category: "Marketing",
    lang: "en",
    price: 399000,
    rating: 4.7,
    ratingCount: 1680,
    studentCount: 9100,
    isPublished: true,
    instructor: instructors[3],
    totalLessons: 30,
    totalDurationMinutes: 560,
    createdAt: "2026-03-15",
  },
  {
    id: 11,
    slug: "deep-learning-with-pytorch",
    title: "Deep Learning with PyTorch",
    description: "Build and train neural networks from scratch — CNNs, transformers and deployment.",
    thumbnailColor: "from-violet-400 to-indigo-500",
    category: "Data Science",
    lang: "en",
    price: 749000,
    rating: 4.8,
    ratingCount: 1390,
    studentCount: 6800,
    isPublished: true,
    instructor: instructors[0],
    totalLessons: 44,
    totalDurationMinutes: 980,
    createdAt: "2026-06-20",
  },
  {
    id: 12,
    slug: "photography-masterclass",
    title: "Photography Masterclass: From Beginner to Pro",
    description: "Master composition, lighting and editing to take stunning photos with any camera.",
    thumbnailColor: "from-slate-400 to-zinc-500",
    category: "Photography",
    lang: "en",
    price: 249000,
    rating: 4.9,
    ratingCount: 3400,
    studentCount: 21500,
    isPublished: true,
    instructor: instructors[1],
    totalLessons: 34,
    totalDurationMinutes: 620,
    createdAt: "2026-02-08",
  },
];

export const categoryTree: import("@/types").CategoryNode[] = [
  { slug: "Development", name: "Development", icon: "Code2", color: "bg-indigo-100 text-indigo-600", courseCount: 1240, description: "Web, mobile and software engineering." },
  { slug: "Design", name: "Design", icon: "Palette", color: "bg-pink-100 text-pink-600", courseCount: 680, description: "UI/UX, graphic and product design." },
  { slug: "Business", name: "Business", icon: "Briefcase", color: "bg-emerald-100 text-emerald-600", courseCount: 540, description: "Entrepreneurship, finance and strategy." },
  { slug: "Marketing", name: "Marketing", icon: "Megaphone", color: "bg-amber-100 text-amber-600", courseCount: 410, description: "Growth, SEO, social and content." },
  { slug: "Data Science", name: "Data Science", icon: "LineChart", color: "bg-sky-100 text-sky-600", courseCount: 390, description: "ML, AI, analytics and Python." },
  { slug: "Photography", name: "Photography", icon: "Camera", color: "bg-violet-100 text-violet-600", courseCount: 220, description: "Shooting, editing and visual storytelling." },
];

export const orders: import("@/types").Order[] = [
  {
    id: "ORD-2026-1042",
    date: "2026-06-10",
    status: "completed",
    paymentMethod: "Visa •••• 4242",
    items: [
      { courseTitle: "Complete Next.js 16 Developer Course", instructor: "Sarah Mitchell", thumbnailColor: "from-indigo-400 to-violet-500", price: 499000 },
      { courseTitle: "Advanced TypeScript Deep Dive", instructor: "Sarah Mitchell", thumbnailColor: "from-fuchsia-400 to-purple-500", price: 349000 },
    ],
    total: 84.98,
  },
  {
    id: "ORD-2026-0871",
    date: "2026-04-22",
    status: "completed",
    paymentMethod: "Mastercard •••• 5511",
    items: [
      { courseTitle: "UI/UX Design Masterclass 2026", instructor: "James Carter", thumbnailColor: "from-emerald-400 to-teal-500", price: 399000 },
    ],
    total: 39.99,
  },
  {
    id: "ORD-2026-0610",
    date: "2026-03-02",
    status: "refunded",
    paymentMethod: "PayPal",
    items: [
      { courseTitle: "Data-Driven Product Management", instructor: "Elena Rodriguez", thumbnailColor: "from-amber-400 to-orange-500", price: 599000 },
    ],
    total: 59.99,
  },
];

export const notifications: import("@/types").Notification[] = [
  { id: 1, type: "course", title: "New lesson published", body: "“Server Actions in depth” was added to Complete Next.js 16 Developer Course.", createdAt: "2026-06-25T09:20:00Z", read: false },
  { id: 2, type: "message", title: "Sarah Mitchell replied", body: "Answered your question in the Routing & Layouts lesson.", createdAt: "2026-06-24T16:05:00Z", read: false },
  { id: 3, type: "promo", title: "Flash sale — 80% off", body: "Top courses are on sale for the next 48 hours. Don't miss out!", createdAt: "2026-06-23T08:00:00Z", read: true },
  { id: 4, type: "system", title: "Certificate ready", body: "Your certificate for Python for Data Science is ready to download.", createdAt: "2026-06-20T11:30:00Z", read: true },
  { id: 5, type: "course", title: "You're on a 7-day streak 🔥", body: "Keep learning today to extend your streak.", createdAt: "2026-06-19T07:00:00Z", read: true },
];

export const enrolledCourses: EnrolledCourse[] = [
  { enrollmentId: 1, course: courses[0], progress: 33, currentLesson: "Nested routes & layouts", lessonsCompleted: 14 },
  { enrollmentId: 2, course: courses[1], progress: 73, currentLesson: "Design Systems", lessonsCompleted: 22 },
  { enrollmentId: 3, course: courses[2], progress: 18, currentLesson: "Metrics that matter", lessonsCompleted: 5 },
];

export const certificates: Certificate[] = [
  { id: 1, courseTitle: "Python for Data Science", issuedAt: "2026-03-02", color: "bg-amber-100" },
  { id: 2, courseTitle: "SQL Fundamentals", issuedAt: "2026-02-11", color: "bg-emerald-100" },
];

export const sampleQuiz: Quiz = {
  id: 1,
  title: "Routing & Layouts Quiz",
  passingScore: 70,
  timeLimitMinutes: 10,
  questions: [
    { id: 1, question: "Which file defines a shared layout in the App Router?", options: ["page.tsx", "layout.tsx", "route.ts", "loading.tsx"], correctIndex: 1 },
    { id: 2, question: "What folder convention creates a dynamic route segment?", options: ["[id]", "(id)", "{id}", "<id>"], correctIndex: 0 },
    { id: 3, question: "Server Components render where by default?", options: ["Browser", "Server", "Edge only", "Web worker"], correctIndex: 1 },
  ],
};
