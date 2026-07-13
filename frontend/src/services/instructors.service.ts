import { api, USE_MOCK } from "@/lib/axios";
import type { Instructor } from "@/types";
import { instructors } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export const instructorsService = {
  async getById(id: string | number): Promise<Instructor> {
    if (USE_MOCK) {
      await delay(250);
      const instructor = instructors.find((i) => i.id === Number(id));
      if (!instructor) throw new Error("Instructor not found");
      return instructor;
    }
    const { data } = await api.get(`/instructors/${id}`);
    return data.instructor;
  },

  async list(): Promise<Instructor[]> {
    if (USE_MOCK) {
      await delay();
      return instructors;
    }
    const { data } = await api.get("/instructors");
    return data.items;
  },
};
