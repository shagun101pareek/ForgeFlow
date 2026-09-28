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

export type GenerationResult = {
  id: string;
  projectId: string;
  prompt: string;
  status: "completed";
  specification: UISpecification;
  files: GeneratedFile[];
  createdAt: string;
};

export type GenerationSummary = {
  id: string;
  prompt: string;
  createdAt: string;
};

export async function startGeneration(input: GenerationPromptInput) {
  const { data } = await api.post<GenerationResult>("/api/v1/generation", input, {
    timeout: 120_000,
  });
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

export async function getGeneration(projectId: string, generationId: string) {
  const { data } = await api.get<GenerationResult>(
    `/api/v1/projects/${projectId}/generations/${generationId}`,
  );
  return data;
}
