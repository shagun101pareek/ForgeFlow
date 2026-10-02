import axios from "axios";

import { api } from "@/lib/axios";
import type { GenerationPromptInput } from "@/lib/validators";

export type GeneratedFile = {
  path: string;
  code: string;
};

export type UISpecification = {
  project: { name: string };
  pages: Array<{
    name: string;
    route: string;
    sections: Array<{ type: string; title: string }>;
  }>;
};

export type GenerationStatus = "queued" | "running" | "completed" | "failed";

export type GenerationResult = {
  id: string;
  projectId: string;
  prompt: string;
  status: GenerationStatus;
  error: string;
  specification: UISpecification;
  files: GeneratedFile[];
  createdAt: string;
};

export type GenerationSummary = {
  id: string;
  prompt: string;
  status: GenerationStatus;
  createdAt: string;
};

export function generationIsActive(status: GenerationStatus | undefined) {
  return status === "queued" || status === "running";
}

export async function enqueueGeneration(input: GenerationPromptInput) {
  const { data } = await api.post<GenerationResult>("/api/v1/generation", input);
  return data;
}

export async function startGeneration(input: GenerationPromptInput) {
  const queued = await enqueueGeneration(input);
  return waitForGeneration(input.projectId, queued.id);
}

export async function waitForGeneration(projectId: string, generationId: string) {
  const deadline = Date.now() + 120_000;
  for (;;) {
    const record = await getGeneration(projectId, generationId);
    if (record.status === "completed") {
      return record;
    }
    if (record.status === "failed") {
      throw new Error(record.error || "Could not generate the interface");
    }
    if (Date.now() >= deadline) {
      throw new Error("Generation timed out");
    }
    await new Promise((resolve) => setTimeout(resolve, 400));
  }
}

export async function saveGeneration(
  projectId: string,
  generationId: string,
  files: GeneratedFile[],
) {
  const { data } = await api.post<GenerationResult>(
    `/api/v1/projects/${projectId}/generations`,
    { generationId, files },
  );
  return data;
}

export async function listGenerations(projectId: string) {
  const { data } = await api.get<GenerationSummary[]>(
    `/api/v1/projects/${projectId}/generations`,
  );
  return data ?? [];
}

export async function getLatestGeneration(projectId: string) {
  try {
    const { data } = await api.get<GenerationResult>(
      `/api/v1/projects/${projectId}/generations/latest`,
    );
    return data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      return null;
    }
    throw error;
  }
}

export async function downloadGeneration(projectId: string, generationId: string) {
  const response = await api.get<Blob>(
    `/api/v1/projects/${projectId}/generations/${generationId}/export`,
    { responseType: "blob" },
  );
  const match = /filename="([^"]+)"/.exec(
    response.headers["content-disposition"] ?? "",
  );
  const filename = match?.[1] ?? "forgeflow.zip";
  const url = URL.createObjectURL(response.data);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

export async function getGeneration(projectId: string, generationId: string) {
  const { data } = await api.get<GenerationResult>(
    `/api/v1/projects/${projectId}/generations/${generationId}`,
  );
  return data;
}
