"use client";

import type { ReactNode } from "react";

import { AuthHydration } from "@/features/auth/auth-hydration";
import { QueryProvider } from "@/providers/query-provider";
import { ThemeProvider } from "@/providers/theme-provider";
import { ToastProvider } from "@/providers/toast-provider";

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider>
      <QueryProvider>
        <AuthHydration />
        {children}
        <ToastProvider />
      </QueryProvider>
    </ThemeProvider>
  );
}
