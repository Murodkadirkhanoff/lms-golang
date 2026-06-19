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

// Normalize the Go API's `{ error: ... }` envelope into thrown Error messages.
api.interceptors.response.use(
  (response) => response,
  (error) => {
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
