import axios from "axios";

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000/v1",
  headers: { "Content-Type": "application/json" },
});

// Attach auth token (stored client-side after login).
api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("token");
    if (token) config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Normalize the API's `{ error: ... }` envelope into thrown Error messages.
// A 401 with a stored token means the session expired: clear it and send the
// user to login with the current page preserved (skipped for the login call
// itself, where 401 just means wrong credentials).
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;
    const url: string = error?.config?.url ?? "";
    if (
      status === 401 &&
      typeof window !== "undefined" &&
      localStorage.getItem("token") &&
      !url.includes("/tokens/authentication")
    ) {
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      if (!window.location.pathname.startsWith("/login")) {
        const next = window.location.pathname + window.location.search;
        window.location.href = `/login?next=${encodeURIComponent(next)}`;
      }
    }

    const payload = error?.response?.data?.error;
    const message =
      typeof payload === "string"
        ? payload
        : payload
          ? Object.values(payload).join(", ")
          : error.message ?? "Something went wrong";
    return Promise.reject(new Error(message));
  },
);

export const USE_MOCK = process.env.NEXT_PUBLIC_USE_MOCK !== "false";
