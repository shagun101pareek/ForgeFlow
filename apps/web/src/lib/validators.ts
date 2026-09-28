import { z } from "zod";

export const loginSchema = z.object({
  email: z.email(),
  password: z.string().min(8),
});

export const signupSchema = loginSchema.extend({
  name: z.string().min(1).max(100),
});

export const createProjectSchema = z.object({
  name: z.string().min(1).max(120),
  description: z.string().max(2000).optional(),
});

export const generationPromptSchema = z.object({
  prompt: z.string().min(1).max(8000),
  projectId: z.string().min(1),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type SignupInput = z.infer<typeof signupSchema>;
export type CreateProjectInput = z.infer<typeof createProjectSchema>;
export type GenerationPromptInput = z.infer<typeof generationPromptSchema>;
