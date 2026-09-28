import { api } from "@/lib/axios";
import type { LoginInput, SignupInput } from "@/lib/validators";

export type AuthResponse = {
  token: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
};

export async function login(input: LoginInput) {
  const { data } = await api.post<AuthResponse>("/api/v1/auth/login", input);
  return data;
}

export async function signup(input: SignupInput) {
  const { data } = await api.post<AuthResponse>("/api/v1/auth/signup", input);
  return data;
}

export async function me() {
  const { data } = await api.get<AuthResponse["user"]>("/api/v1/auth/me");
  return data;
}
