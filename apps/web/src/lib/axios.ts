import axios from "axios";

import { useAuthStore } from "@/store/auth";

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080",
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 30_000,
});

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = window.localStorage.getItem("forgeflow_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const url = error.config?.url ?? "";
      const isCredentialAttempt =
        url.includes("/auth/login") || url.includes("/auth/signup");
      if (!isCredentialAttempt) {
        useAuthStore.getState().clearAuth();
      }
    }
    return Promise.reject(error);
  },
);
