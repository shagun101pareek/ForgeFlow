import { api } from "@/lib/axios";

export type UserProfile = {
  id: string;
  email: string;
  name: string;
};

export async function getCurrentUser() {
  const { data } = await api.get<UserProfile>("/api/v1/users/me");
  return data;
}

export async function updateCurrentUser(input: Partial<Pick<UserProfile, "name">>) {
  const { data } = await api.patch<UserProfile>("/api/v1/users/me", input);
  return data;
}
