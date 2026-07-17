export interface AdminUser {
  id: number;
  name: string;
  email: string;
  role: "student" | "instructor" | "admin";
  avatarColor: string;
  joinedAt: string;
  status: "active" | "suspended";
  coursesCreated: number;
  coursesEnrolled: number;
}

export const adminUsers: AdminUser[] = [
  { id: 1, name: "Amir Karimov", email: "amir@mail.com", role: "student", avatarColor: "bg-indigo-200", joinedAt: "2026-06-12", status: "active", coursesCreated: 0, coursesEnrolled: 8 },
  { id: 2, name: "Sarah Mitchell", email: "sarah@mail.com", role: "student", avatarColor: "bg-amber-200", joinedAt: "2026-05-03", status: "active", coursesCreated: 14, coursesEnrolled: 3 },
  { id: 3, name: "John Doe", email: "john@mail.com", role: "student", avatarColor: "bg-rose-200", joinedAt: "2026-04-21", status: "suspended", coursesCreated: 1, coursesEnrolled: 2 },
  { id: 4, name: "Laura Bennett", email: "laura@mail.com", role: "student", avatarColor: "bg-emerald-200", joinedAt: "2026-04-02", status: "active", coursesCreated: 5, coursesEnrolled: 12 },
  { id: 5, name: "David Park", email: "david@mail.com", role: "student", avatarColor: "bg-sky-200", joinedAt: "2026-03-15", status: "active", coursesCreated: 9, coursesEnrolled: 1 },
  { id: 6, name: "Elena Rodriguez", email: "elena@mail.com", role: "student", avatarColor: "bg-fuchsia-200", joinedAt: "2026-02-28", status: "active", coursesCreated: 6, coursesEnrolled: 4 },
];
