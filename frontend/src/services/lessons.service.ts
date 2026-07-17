import { api, USE_MOCK } from "@/lib/axios";

export interface LessonQuestion {
  id: number;
  createdAt: string;
  user: string;
  question: string;
}

const mockQuestions: LessonQuestion[] = [];

export const lessonsService = {
  async listQuestions(lessonId: number): Promise<LessonQuestion[]> {
    if (USE_MOCK) {
      return mockQuestions.filter(() => true);
    }
    const { data } = await api.get(`/lessons/${lessonId}/questions`);
    return data.items;
  },

  async askQuestion(lessonId: number, question: string): Promise<LessonQuestion> {
    if (USE_MOCK) {
      const q = { id: Date.now(), createdAt: new Date().toISOString(), user: "You", question };
      mockQuestions.unshift(q);
      return q;
    }
    const { data } = await api.post(`/lessons/${lessonId}/questions`, { question });
    return data.question;
  },
};
