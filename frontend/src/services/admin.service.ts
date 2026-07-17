import { api, USE_MOCK } from "@/lib/axios";
import { adminUsers, type AdminUser } from "./mock/users";
import type { Paginated } from "@/types";

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

  async getUsers(page = 1, pageSize = 20): Promise<Paginated<AdminUser>> {
    if (USE_MOCK) {
      await delay();
      const start = (page - 1) * pageSize;
      return { items: adminUsers.slice(start, start + pageSize), page, pageSize, total: adminUsers.length };
    }
    const { data } = await api.get("/admin/users", { params: { page, pageSize } });
    return data;
  },

  /** PATCH /admin/users/{id}/role — admin/instructor/student tayinlash. */
  async updateRole(id: number, role: AdminUser["role"]): Promise<void> {
    if (USE_MOCK) {
      await delay(200);
      const u = adminUsers.find((x) => x.id === id);
      if (u) u.role = role;
      return;
    }
    await api.patch(`/admin/users/${id}/role`, { role });
  },
};

export type { AdminUser };
