"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { Skeleton } from "@/components/ui/skeleton";
import { me } from "@/services/auth";
import { useAuthStore } from "@/store/auth";

export function RequireAuth({ children }: { children: ReactNode }) {
  const router = useRouter();
  const token = useAuthStore((state) => state.token);
  const hasHydrated = useAuthStore((state) => state.hasHydrated);
  const setAuth = useAuthStore((state) => state.setAuth);
  const clearAuth = useAuthStore((state) => state.clearAuth);

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
    if (!hasHydrated) {
      return;
    }
    if (!token || profile.isError) {
      if (profile.isError) {
        clearAuth();
      }
      router.replace("/login");
    }
  }, [clearAuth, hasHydrated, profile.isError, router, token]);

  if (!hasHydrated || !token || profile.isPending) {
    return (
      <main className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-3 px-6 py-24">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
      </main>
    );
  }

  if (profile.isError || !profile.data) {
    return (
      <main className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-3 px-6 py-24">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-4 w-full" />
      </main>
    );
  }

  return children;
}
