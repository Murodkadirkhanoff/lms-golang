import { api, USE_MOCK } from "@/lib/axios";
import type { Certificate, EnrolledCourse, TeachingStats } from "@/types";
import { certificates, enrolledCourses } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export interface DashboardStats {
  enrolled: number;
  inProgress: number;
  completed: number;
  certificates: number;
}

export const dashboardService = {
  async getStats(): Promise<DashboardStats> {
    if (USE_MOCK) {
      await delay(300);
      return { enrolled: 8, inProgress: 3, completed: 5, certificates: certificates.length };
    }
    const { data } = await api.get("/me/stats");
    return data;
  },

  async getEnrolled(): Promise<EnrolledCourse[]> {
    if (USE_MOCK) {
      await delay();
      return enrolledCourses;
    }
    const { data } = await api.get("/me/courses");
    return data.items;
  },

  async getCertificates(): Promise<Certificate[]> {
    if (USE_MOCK) {
      await delay();
      return certificates;
    }
    const { data } = await api.get("/me/certificates");
    return data.items;
  },

  // Studio (instruktor) ko'rsatkichlari — GET /me/teaching/stats.
  async getTeachingStats(): Promise<TeachingStats> {
    if (USE_MOCK) {
      await delay(300);
      return {
        totalRevenue: 48250000,
        monthlyRevenue: [
          { month: "2026-02", revenue: 4500000 },
          { month: "2026-03", revenue: 6000000 },
          { month: "2026-04", revenue: 5200000 },
          { month: "2026-05", revenue: 7800000 },
          { month: "2026-06", revenue: 6800000 },
          { month: "2026-07", revenue: 10000000 },
        ],
        totalStudents: 12840,
        activeStudents: 8420,
        publishedCourses: 14,
        draftCourses: 2,
        avgRating: 4.8,
        avgCompletion: 64,
        avgQuizScore: 81,
        engagement: [
          { courseId: 1, title: "Complete Next.js 16 Course", students: 2340, completion: 68 },
          { courseId: 2, title: "React Patterns & Performance", students: 1820, completion: 54 },
          { courseId: 3, title: "Advanced TypeScript Deep Dive", students: 1440, completion: 72 },
        ],
      };
    }
    const { data } = await api.get("/me/teaching/stats");
    return data;
  },
};
