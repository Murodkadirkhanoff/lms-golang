import { api, USE_MOCK } from "@/lib/axios";
import type { Order } from "@/types";
import { orders } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export const ordersService = {
  async list(): Promise<Order[]> {
    if (USE_MOCK) {
      await delay();
      return orders;
    }
    const { data } = await api.get("/me/orders");
    return data.items;
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
};
