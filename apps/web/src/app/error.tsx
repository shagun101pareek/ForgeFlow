"use client";

import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Wordmark } from "@/components/wordmark";

export default function AppError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="flex flex-1 flex-col">
      <header className="px-4 py-4 sm:px-6">
        <Wordmark />
      </header>
      <div className="mx-auto flex w-full max-w-lg flex-1 flex-col justify-center gap-4 px-6 py-16">
        <p className="text-sm font-medium tracking-[0.18em] text-primary uppercase">
          Error
        </p>
        <h1 className="font-heading text-4xl font-medium tracking-tight">
          Something went wrong
        </h1>
        <p className="text-sm leading-6 text-muted-foreground">
          {error.message || "The page hit an unexpected error."}
        </p>
        <div className="flex flex-wrap gap-2 pt-2">
          <Button type="button" onClick={reset}>
            Try again
          </Button>
          <Button
            variant="outline"
            nativeButton={false}
            render={<Link href="/dashboard" />}
          >
            Projects
          </Button>
        </div>
      </div>
    </main>
  );
}
