import { api, USE_MOCK } from "@/lib/axios";
import type { Quiz } from "@/types";
import { sampleQuiz } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export const quizService = {
  async getById(id: string | number): Promise<Quiz> {
    if (USE_MOCK) {
      await delay();
      return sampleQuiz;
    }
    const { data } = await api.get(`/quizzes/${id}`);
    return data.quiz;
  },
};
