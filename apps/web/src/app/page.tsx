import Link from "next/link";

import { Wordmark } from "@/components/wordmark";
import { HomeActions } from "@/features/auth/home-actions";
import { examplePrompts } from "@/features/generation/example-prompts";

export default function Home() {
  return (
    <main className="flex flex-1 flex-col">
      <header className="flex items-center px-4 py-4 sm:px-6">
        <Wordmark />
      </header>
      <div className="mx-auto grid w-full max-w-6xl flex-1 items-center gap-12 px-6 py-12 lg:grid-cols-[1.1fr_0.9fr] lg:py-20">
        <div>
          <p className="text-sm font-medium tracking-[0.18em] text-primary uppercase">
            Prompt to preview
          </p>
          <h1 className="mt-4 font-heading text-5xl leading-[1.05] font-medium tracking-tight text-balance sm:text-6xl">
            An interface you can click through.
          </h1>
          <p className="mt-5 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
            Describe a page, then open the preview, edit the source, and
            download the result.
          </p>
          <div className="mt-8">
            <HomeActions />
          </div>
        </div>
        <PreviewFrame />
      </div>
      <ul className="mx-auto grid w-full max-w-6xl gap-3 px-6 pb-16 sm:grid-cols-2 lg:grid-cols-4">
        {examplePrompts.map((example) => (
          <li key={example.id}>
            <Link
              href="/signup"
              className={`flex h-full flex-col gap-3 rounded-2xl p-4 transition-transform hover:-translate-y-0.5 ${example.tone}`}
            >
              <span className="font-heading text-xl">{example.label}</span>
              <span className="text-sm leading-6 opacity-80">{example.text}</span>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}

function PreviewFrame() {
  return (
    <div className="relative">
      <div
        aria-hidden="true"
        className="absolute -top-8 -left-6 size-32 rounded-full bg-primary/20 blur-3xl"
      />
      <div
        aria-hidden="true"
        className="absolute -right-4 -bottom-8 size-28 rounded-full bg-teal-400/30 blur-3xl"
      />
      <div className="relative overflow-hidden rounded-2xl bg-card shadow-xl ring-1 ring-foreground/10">
        <div className="flex items-center gap-2 border-b px-4 py-3">
          <span className="size-2 rounded-full bg-primary" />
          <span className="size-2 rounded-full bg-foreground/15" />
          <span className="size-2 rounded-full bg-foreground/15" />
          <span className="ml-3 h-2 w-28 rounded-full bg-muted" />
        </div>
        <div className="space-y-4 p-6">
          <p className="text-xs font-medium tracking-[0.16em] text-primary uppercase">
            Notes
          </p>
          <p className="font-heading text-3xl leading-tight">
            Keep the work moving.
          </p>
          <div className="h-2 w-4/5 rounded-full bg-muted" />
          <div className="flex gap-2 pt-1">
            <div className="h-8 w-24 rounded-lg bg-primary" />
            <div className="h-8 w-24 rounded-lg bg-muted" />
          </div>
          <div className="grid grid-cols-3 gap-2 pt-2">
            <div className="h-16 rounded-xl bg-orange-100 dark:bg-orange-950/40" />
            <div className="h-16 rounded-xl bg-teal-100 dark:bg-teal-950/40" />
            <div className="h-16 rounded-xl bg-rose-100 dark:bg-rose-950/40" />
          </div>
        </div>
      </div>
    </div>
  );
}
