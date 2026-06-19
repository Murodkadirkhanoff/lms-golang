import { api, USE_MOCK } from "@/lib/axios";
import type { Course, CourseQuery, Paginated } from "@/types";
import { courses } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

// Mirrors the `lessons` table (duration in seconds; position from order).
export interface CreateLessonInput {
  title: string;
  contentUrl: string;
  durationSeconds: number;
  price: number;
  isFree: boolean;
}

// Mirrors the `modules` table (position from order).
export interface CreateModuleInput {
  title: string;
  lessons: CreateLessonInput[];
}

// Mirrors the `courses` table columns. instructor_id is set server-side from
// the authenticated user, so it is not part of this input.
export interface CreateCourseInput {
  title: string;
  description: string;
  categoryId: number | null;
  lang: "uz" | "ru" | "en";
  price: number;
  isPublished: boolean;
  modules: CreateModuleInput[];
}

export const coursesService = {
  async list(query: CourseQuery = {}): Promise<Paginated<Course>> {
    if (USE_MOCK) {
      await delay();
      const { search = "", category, sort = "popular", page = 1, pageSize = 8 } = query;
      let items = courses.filter((c) => c.isPublished);

      if (search) {
        const q = search.toLowerCase();
        items = items.filter(
          (c) => c.title.toLowerCase().includes(q) || c.description.toLowerCase().includes(q),
        );
      }
      if (category && category !== "All") {
        items = items.filter((c) => c.category === category);
      }

      items = [...items].sort((a, b) => {
        switch (sort) {
          case "newest":
            return +new Date(b.createdAt) - +new Date(a.createdAt);
          case "price-asc":
            return a.price - b.price;
          case "price-desc":
            return b.price - a.price;
          default:
            return b.studentCount - a.studentCount;
        }
      });

      const total = items.length;
      const start = (page - 1) * pageSize;
      return { items: items.slice(start, start + pageSize), page, pageSize, total };
    }

    const { data } = await api.get("/courses", { params: query });
    return data;
  },

  async getBySlug(slug: string): Promise<Course> {
    if (USE_MOCK) {
      await delay();
      const course = courses.find((c) => c.slug === slug);
      if (!course) throw new Error("Course not found");
      return course;
    }
    const { data } = await api.get(`/courses/${slug}`);
    return data.course;
  },

  async getById(id: string | number): Promise<Course> {
    if (USE_MOCK) {
      await delay();
      const course = courses.find((c) => c.id === Number(id)) ?? courses[0];
      return course;
    }
    const { data } = await api.get(`/courses/${id}`);
    return data.course;
  },

  async getPopular(limit = 4): Promise<Course[]> {
    if (USE_MOCK) {
      await delay(300);
      return [...courses].sort((a, b) => b.studentCount - a.studentCount).slice(0, limit);
    }
    const { data } = await api.get("/courses", { params: { sort: "popular", pageSize: limit } });
    return data.items;
  },

  async create(input: CreateCourseInput): Promise<Course> {
    if (USE_MOCK) {
      await delay(600);
      return {
        ...courses[0],
        id: Date.now(),
        slug: input.title.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, ""),
        title: input.title,
        description: input.description,
        lang: input.lang,
        price: input.price,
        isPublished: input.isPublished,
      };
    }
    // Maps to the Go backend columns (snake_case). instructor_id is set by the
    // server from the authenticated user. Modules/lessons carry their position
    // from array order, matching the `modules`/`lessons` tables.
    const { data } = await api.post("/courses", {
      title: input.title,
      description: input.description,
      category_id: input.categoryId,
      lang: input.lang,
      price: input.price,
      is_published: input.isPublished,
      modules: input.modules.map((m, mi) => ({
        title: m.title,
        position: mi,
        lessons: m.lessons.map((l, li) => ({
          title: l.title,
          content_url: l.contentUrl,
          duration_seconds: l.durationSeconds,
          position: li,
          price: l.price,
          is_free: l.isFree,
        })),
      })),
    });
    return data.course;
  },
};
