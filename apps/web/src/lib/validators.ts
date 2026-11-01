import { z } from "zod";

export const loginSchema = z.object({
  email: z.email("Enter a valid email address"),
  password: z.string().min(8, "Use at least 8 characters"),
});

export const signupSchema = loginSchema.extend({
  name: z
    .string()
    .min(1, "Enter your name")
    .max(100, "Name must be 100 characters or fewer"),
});

export const createProjectSchema = z.object({
  name: z
    .string()
    .min(1, "Enter a project name")
    .max(120, "Name must be 120 characters or fewer"),
  description: z
    .string()
    .max(2000, "Description must be 2000 characters or fewer")
    .optional(),
});

export const generationPromptSchema = z.object({
  prompt: z
    .string()
    .min(1, "Enter a prompt")
    .max(8000, "Prompt must be 8000 characters or fewer"),
  projectId: z.string().min(1, "Project is required"),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type SignupInput = z.infer<typeof signupSchema>;
export type CreateProjectInput = z.infer<typeof createProjectSchema>;
export type GenerationPromptInput = z.infer<typeof generationPromptSchema>;
