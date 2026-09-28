import type { Metadata } from "next";

import { RequireAuth } from "@/features/auth/require-auth";
import { DashboardPage } from "@/features/dashboard/dashboard-page";

export const metadata: Metadata = {
  title: "Dashboard · ForgeFlow",
};

export default function DashboardRoute() {
  return (
    <RequireAuth>
      <DashboardPage />
    </RequireAuth>
  );
}
