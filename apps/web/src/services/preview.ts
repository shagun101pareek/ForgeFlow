import { api } from "@/lib/axios";

export type PreviewSession = {
  id: string;
  url: string;
  status: "ready" | "building" | "error";
};

export async function createPreview(projectId: string) {
  const { data } = await api.post<PreviewSession>("/api/v1/preview", {
    projectId,
  });
  return data;
}

export async function getPreview(id: string) {
  const { data } = await api.get<PreviewSession>(`/api/v1/preview/${id}`);
  return data;
}
