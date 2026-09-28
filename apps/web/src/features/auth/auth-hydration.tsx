"use client";

import { useEffect } from "react";

import { useAuthStore } from "@/store/auth";

export function AuthHydration() {
  useEffect(() => {
    const finish = () => {
      const { token } = useAuthStore.getState();
      if (token) {
        window.localStorage.setItem("forgeflow_token", token);
      }
      useAuthStore.getState().setHasHydrated(true);
    };

    if (useAuthStore.persist.hasHydrated()) {
      finish();
    }

    return useAuthStore.persist.onFinishHydration(() => {
      finish();
    });
  }, []);

  return null;
}
