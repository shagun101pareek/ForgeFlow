import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Wordmark } from "@/components/wordmark";

export default function NotFound() {
  return (
    <main className="flex flex-1 flex-col">
      <header className="px-4 py-4 sm:px-6">
        <Wordmark />
      </header>
      <div className="mx-auto flex w-full max-w-lg flex-1 flex-col justify-center gap-4 px-6 py-16">
        <p className="text-sm font-medium tracking-[0.18em] text-primary uppercase">
          404
        </p>
        <h1 className="font-heading text-4xl font-medium tracking-tight">
          Page not found
        </h1>
        <p className="text-sm leading-6 text-muted-foreground">
          That route is not part of ForgeFlow. Head back to the projects list or
          the homepage.
        </p>
        <div className="flex flex-wrap gap-2 pt-2">
          <Button nativeButton={false} render={<Link href="/dashboard" />}>
            Projects
          </Button>
          <Button
            variant="outline"
            nativeButton={false}
            render={<Link href="/" />}
          >
            Home
          </Button>
        </div>
      </div>
    </main>
  );
}
