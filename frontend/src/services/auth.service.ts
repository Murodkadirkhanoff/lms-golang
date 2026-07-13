import { api, USE_MOCK } from "@/lib/axios";
import type { User } from "@/types";

const delay = (ms = 600) => new Promise((r) => setTimeout(r, ms));

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  name: string;
  email: string;
  password: string;
}

export interface AuthResult {
  user: User;
  token: string;
}

const USER_KEY = "user";

function persist(result: AuthResult) {
  if (typeof window === "undefined") return;
  localStorage.setItem("token", result.token);
  localStorage.setItem(USER_KEY, JSON.stringify(result.user));
}

/** The signed-in user cached client-side, or null when logged out. */
export function getStoredUser(): User | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export const authService = {
  async login(input: LoginInput): Promise<AuthResult> {
    if (USE_MOCK) {
      await delay();
      const result: AuthResult = {
        token: "mock-token",
        user: { id: 1, name: "Amir Karimov", email: input.email, createdAt: "2026-06-12" },
      };
      persist(result);
      return result;
    }
    const { data } = await api.post("/tokens/authentication", input);
    persist(data);
    return data;
  },

  async register(input: RegisterInput): Promise<AuthResult> {
    if (USE_MOCK) {
      await delay();
      const result: AuthResult = {
        token: "mock-token",
        user: { id: Date.now(), name: input.name, email: input.email, createdAt: new Date().toISOString() },
      };
      persist(result);
      return result;
    }
    const { data } = await api.post("/users", input);
    return data;
  },

  async forgotPassword(email: string): Promise<{ message: string }> {
    if (USE_MOCK) {
      await delay();
      return { message: `Password reset link sent to ${email}` };
    }
    const { data } = await api.post("/tokens/password-reset", { email });
    return data;
  },

  async resetPassword(password: string, token?: string): Promise<{ message: string }> {
    if (USE_MOCK) {
      await delay();
      return { message: "Your password has been reset" };
    }
    const { data } = await api.put("/users/password", { password, token });
    return data;
  },

  logout() {
    if (typeof window === "undefined") return;
    localStorage.removeItem("token");
    localStorage.removeItem(USER_KEY);
  },
};
