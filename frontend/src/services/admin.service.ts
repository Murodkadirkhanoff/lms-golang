import { api, USE_MOCK } from "@/lib/axios";
import { adminUsers, type AdminUser } from "./mock/users";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export interface AdminStats {
  totalUsers: number;
  totalCourses: number;
  revenue: number;
  activeInstructors: number;
}

export const adminService = {
  async getStats(): Promise<AdminStats> {
    if (USE_MOCK) {
      await delay(300);
      return { totalUsers: 320480, totalCourses: 2540, revenue: 1_240_000, activeInstructors: 850 };
    }
    const { data } = await api.get("/admin/stats");
    return data;
  },

  async getUsers(): Promise<AdminUser[]> {
    if (USE_MOCK) {
      await delay();
      return adminUsers;
    }
    const { data } = await api.get("/admin/users");
    return data.items;
  },
};

export type { AdminUser };
