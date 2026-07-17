import { api, USE_MOCK } from "@/lib/axios";
import type { Quiz, QuizAttempt } from "@/types";
import { sampleQuiz } from "./mock/data";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

const mockAttempts: QuizAttempt[] = [
  { id: 1, createdAt: "2026-06-14T10:00:00Z", score: 80 },
  { id: 2, createdAt: "2026-06-10T10:00:00Z", score: 60 },
];

// Studio quiz builder yuboradigan shakl (PUT /courses/{id}/quiz, snake_case).
export interface UpsertQuizInput {
  title: string;
  passingScore: number;
  timeLimitMinutes: number;
  questions: { question: string; options: string[]; correctIndex: number }[];
}

// {id} — kurs id'si (har kursga bitta quiz).
export const quizService = {
  async getById(id: string | number): Promise<Quiz> {
    if (USE_MOCK) {
      await delay();
      return sampleQuiz;
    }
    const { data } = await api.get(`/quizzes/${id}`);
    return data.quiz;
  },

  /** Mavjud bo'lmasa null — builder bo'sh formadan boshlaydi. */
  async getForCourse(courseId: string | number): Promise<Quiz | null> {
    if (USE_MOCK) {
      await delay();
      return sampleQuiz;
    }
    try {
      const { data } = await api.get(`/quizzes/${courseId}`);
      return data.quiz;
    } catch {
      return null;
    }
  },

  async upsert(courseId: string | number, input: UpsertQuizInput): Promise<Quiz> {
    if (USE_MOCK) {
      await delay();
      return sampleQuiz;
    }
    const { data } = await api.put(`/courses/${courseId}/quiz`, {
      title: input.title,
      passing_score: input.passingScore,
      time_limit_minutes: input.timeLimitMinutes,
      questions: input.questions.map((q) => ({
        question: q.question,
        options: q.options,
        correct_index: q.correctIndex,
      })),
    });
    return data.quiz;
  },

  async listAttempts(id: string | number): Promise<QuizAttempt[]> {
    if (USE_MOCK) {
      await delay(200);
      return mockAttempts;
    }
    const { data } = await api.get(`/quizzes/${id}/attempts`);
    return data.attempts;
  },

  async submitAttempt(id: string | number, score: number): Promise<QuizAttempt> {
    if (USE_MOCK) {
      await delay(200);
      const attempt = { id: Date.now(), createdAt: new Date().toISOString(), score };
      mockAttempts.unshift(attempt);
      return attempt;
    }
    const { data } = await api.post(`/quizzes/${id}/attempts`, { score });
    return data.attempt;
  },
};
