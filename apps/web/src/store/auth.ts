import { create } from "zustand";
import { persist } from "zustand/middleware";

type User = {
  id: string;
  email: string;
  name: string;
};

type AuthState = {
  token: string | null;
  user: User | null;
  hasHydrated: boolean;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  setHasHydrated: (value: boolean) => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      hasHydrated: false,
      setAuth: (token, user) => {
        if (typeof window !== "undefined") {
          window.localStorage.setItem("forgeflow_token", token);
        }
        set({ token, user });
      },
      clearAuth: () => {
        if (typeof window !== "undefined") {
          window.localStorage.removeItem("forgeflow_token");
        }
        set({ token: null, user: null });
      },
      setHasHydrated: (value) => set({ hasHydrated: value }),
    }),
    {
      name: "forgeflow-auth",
      partialize: (state) => ({
        token: state.token,
        user: state.user,
      }),
      onRehydrateStorage: () => (state) => {
        if (state?.token && typeof window !== "undefined") {
          window.localStorage.setItem("forgeflow_token", state.token);
        }
        state?.setHasHydrated(true);
      },
    },
  ),
);
