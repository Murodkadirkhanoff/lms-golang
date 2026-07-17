import { api, USE_MOCK } from "@/lib/axios";
import type { Order, Paginated } from "@/types";
import { orders } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export const ordersService = {
  async list(page = 1, pageSize = 20): Promise<Paginated<Order>> {
    if (USE_MOCK) {
      await delay();
      const start = (page - 1) * pageSize;
      return { items: orders.slice(start, start + pageSize), page, pageSize, total: orders.length };
    }
    const { data } = await api.get("/me/orders", { params: { page, pageSize } });
    return data;
  },

  async getById(id: string): Promise<Order> {
    if (USE_MOCK) {
      await delay(250);
      const order = orders.find((o) => o.id === id);
      if (!order) throw new Error("Order not found");
      return order;
    }
    const { data } = await api.get(`/me/orders/${id}`);
    return data.order;
  },

  // POST /me/orders — narxlar serverda hisoblanadi, to'lov hozircha mock
  // (buyurtma darhol "paid" bo'lib, kurslarga kirish ochiladi).
  async checkout(items: CheckoutItem[], paymentMethod: string): Promise<Order> {
    if (USE_MOCK) {
      await delay(800);
      return orders[0];
    }
    const { data } = await api.post("/me/orders", {
      items: items.map((i) => ({ course_id: i.courseId ?? null, lesson_id: i.lessonId ?? null })),
      payment_method: paymentMethod,
    });
    return data.order;
  },
};

export interface CheckoutItem {
  courseId?: number;
  lessonId?: number;
}
