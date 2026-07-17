import { api, USE_MOCK } from "@/lib/axios";
import type { Notification } from "@/types";
import { notifications } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export const notificationsService = {
  async list(): Promise<Notification[]> {
    if (USE_MOCK) {
      await delay();
      return notifications;
    }
    const { data } = await api.get("/me/notifications");
    return data.items;
  },

  async markAllRead(): Promise<void> {
    if (USE_MOCK) {
      await delay(200);
      return;
    }
    await api.post("/me/notifications/read-all");
  },

  async markRead(id: number): Promise<void> {
    if (USE_MOCK) {
      await delay(200);
      const n = notifications.find((x) => x.id === id);
      if (n) n.read = true;
      return;
    }
    await api.post(`/me/notifications/${id}/read`);
  },
};
