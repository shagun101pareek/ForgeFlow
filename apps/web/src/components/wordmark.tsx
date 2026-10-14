import Link from "next/link";

export function Wordmark({ href = "/" }: { href?: string }) {
  return (
    <Link
      href={href}
      className="inline-flex items-center gap-2 text-sm font-semibold tracking-tight text-foreground"
    >
      <span aria-hidden="true" className="relative flex size-5 items-center justify-center">
        <span className="absolute size-3.5 rotate-45 rounded-[3px] bg-primary" />
        <span className="absolute size-1.5 rounded-full bg-primary-foreground" />
      </span>
      ForgeFlow
    </Link>
  );
}
