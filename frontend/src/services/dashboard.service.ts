import { api, USE_MOCK } from "@/lib/axios";
import type { Certificate, EnrolledCourse } from "@/types";
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
};
