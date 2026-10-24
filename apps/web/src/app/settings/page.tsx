import type { Metadata } from "next";

import { RequireAuth } from "@/features/auth/require-auth";
import { SettingsPage } from "@/features/settings/settings-page";

export const metadata: Metadata = {
  title: "Settings · ForgeFlow",
};

export default function SettingsRoute() {
  return (
    <RequireAuth>
      <SettingsPage />
    </RequireAuth>
  );
}
