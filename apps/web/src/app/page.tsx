import { HomeActions } from "@/features/auth/home-actions";

export default function Home() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-24">
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-6 text-center">
        <p className="text-sm font-medium tracking-[0.2em] text-muted-foreground uppercase">
          ForgeFlow
        </p>
        <h1 className="text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
          Generate production-ready interactive UI prototypes using AI.
        </h1>
        <p className="max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
          Create an account, open a project, and describe the interface you
          want to build.
        </p>
        <HomeActions />
      </div>
    </main>
  );
}
