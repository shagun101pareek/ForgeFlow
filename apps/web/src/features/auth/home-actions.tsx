"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/store/auth";

export function HomeActions() {
  const router = useRouter();
  const token = useAuthStore((state) => state.token);
  const hasHydrated = useAuthStore((state) => state.hasHydrated);

  useEffect(() => {
    if (hasHydrated && token) {
      router.replace("/dashboard");
    }
  }, [hasHydrated, router, token]);

  return (
    <div className="flex flex-wrap items-center justify-center gap-3">
      <Button render={<Link href="/signup" />}>Create an account</Button>
      <Button variant="outline" render={<Link href="/login" />}>
        Log in
      </Button>
    </div>
  );
}
