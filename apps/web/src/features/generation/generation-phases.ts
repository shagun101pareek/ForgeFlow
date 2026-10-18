export const GENERATION_PHASES = [
  "Understanding prompt",
  "Designing layout",
  "Writing React",
  "Preparing preview",
] as const;

export type GenerationPhase = (typeof GENERATION_PHASES)[number];

const PHASE_MS = 1800;

/** Client-side phase while a job is queued/running. No API changes required. */
export function generationPhase(
  status: string | undefined,
  startedAt: number | null,
  now = Date.now(),
): { label: GenerationPhase; index: number; progress: number } {
  if (status === "queued") {
    return { label: GENERATION_PHASES[0], index: 0, progress: 8 };
  }

  const elapsed = startedAt ? Math.max(0, now - startedAt) : 0;
  const index = Math.min(
    GENERATION_PHASES.length - 1,
    Math.floor(elapsed / PHASE_MS),
  );
  const within = elapsed - index * PHASE_MS;
  const segment = Math.min(1, within / PHASE_MS);
  const progress = Math.min(
    92,
    Math.round(((index + segment) / GENERATION_PHASES.length) * 100),
  );

  return {
    label: GENERATION_PHASES[index],
    index,
    progress: status === "running" ? Math.max(18, progress) : progress,
  };
}
