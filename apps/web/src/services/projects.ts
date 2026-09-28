import { api } from "@/lib/axios";
import type { CreateProjectInput } from "@/lib/validators";
import type { Project } from "@/store/project";

export async function listProjects() {
  const { data } = await api.get<Project[]>("/api/v1/projects");
  return data;
}

export async function getProject(id: string) {
  const { data } = await api.get<Project>(`/api/v1/projects/${id}`);
  return data;
}

export async function createProject(input: CreateProjectInput) {
  const { data } = await api.post<Project>("/api/v1/projects", input);
  return data;
}

export type UpdateProjectInput = {
  name?: string;
  description?: string;
};

export async function updateProject(id: string, input: UpdateProjectInput) {
  const { data } = await api.patch<Project>(`/api/v1/projects/${id}`, input);
  return data;
}

export async function deleteProject(id: string) {
  await api.delete(`/api/v1/projects/${id}`);
}
