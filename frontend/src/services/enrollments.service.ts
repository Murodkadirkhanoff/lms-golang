import { api, USE_MOCK } from "@/lib/axios";

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

export interface Enrollment {
  id: number;
  createdAt: string;
  userId: number;
  courseId: number;
}

export const enrollmentsService = {
  // Faqat bepul kurslar uchun; pullik kurslar checkout orqali.
  async enroll(courseId: number): Promise<Enrollment> {
    if (USE_MOCK) {
      await delay();
      return { id: Date.now(), createdAt: new Date().toISOString(), userId: 1, courseId };
    }
    const { data } = await api.post(`/courses/${courseId}/enroll`);
    return data.enrollment;
  },

  async updateProgress(enrollmentId: number, lessonId: number, completed: boolean): Promise<void> {
    if (USE_MOCK) {
      await delay(200);
      return;
    }
    await api.patch(`/enrollments/${enrollmentId}/progress`, { lesson_id: lessonId, completed });
  },
};
