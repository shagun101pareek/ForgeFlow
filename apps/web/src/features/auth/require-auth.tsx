"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { me } from "@/services/auth";
import { useAuthStore } from "@/store/auth";

export function RequireAuth({ children }: { children: ReactNode }) {
  const router = useRouter();
  const token = useAuthStore((state) => state.token);
  const hasHydrated = useAuthStore((state) => state.hasHydrated);
  const setAuth = useAuthStore((state) => state.setAuth);

  const profile = useQuery({
    queryKey: ["me", token],
    queryFn: me,
    enabled: hasHydrated && Boolean(token),
    retry: false,
  });

  useEffect(() => {
    if (profile.data && token) {
      setAuth(token, profile.data);
    }
  }, [profile.data, setAuth, token]);

  useEffect(() => {
    if (!hasHydrated || token) {
      return;
    }
    router.replace("/login");
  }, [hasHydrated, router, token]);

  if (!hasHydrated || !token) {
    return (
      <main className="flex flex-1 items-center justify-center px-6 py-24 text-sm text-muted-foreground">
        Loading account…
      </main>
    );
  }

  return children;
}
