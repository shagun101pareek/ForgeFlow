import { create } from "zustand";

import type { GenerationResult, UISpecification } from "@/services/generation";

type EditorFile = {
  path: string;
  code: string;
  language?: string;
};

export type EditorPage = {
  id: string;
  name: string;
  route: string;
};

export const defaultEditorPages: EditorPage[] = [
  { id: "home", name: "Home", route: "/" },
  { id: "about", name: "About", route: "/about" },
  { id: "pricing", name: "Pricing", route: "/pricing" },
];

type EditorState = {
  files: EditorFile[];
  activeFilePath: string | null;
  pages: EditorPage[];
  activePageId: string;
  prompt: string;
  specification: UISpecification | null;
  generationId: string | null;
  generationError: string | null;
  savedFiles: string;
  setFiles: (files: EditorFile[]) => void;
  setActiveFile: (path: string | null) => void;
  updateFile: (path: string, code: string) => void;
  setActivePage: (id: string) => void;
  setPrompt: (prompt: string) => void;
  loadGeneration: (generation: GenerationResult) => void;
  resetWorkspace: () => void;
};

export const useEditorStore = create<EditorState>((set) => ({
  files: [],
  activeFilePath: null,
  pages: defaultEditorPages,
  activePageId: defaultEditorPages[0].id,
  prompt: "",
  specification: null,
  generationId: null,
  generationError: null,
  savedFiles: "[]",
  setFiles: (files) =>
    set({
      files,
      activeFilePath: files[0]?.path ?? null,
    }),
  setActiveFile: (path) => set({ activeFilePath: path }),
  updateFile: (path, code) =>
    set((state) => ({
      files: state.files.map((file) =>
        file.path === path ? { ...file, code } : file,
      ),
    })),
  setActivePage: (id) => set({ activePageId: id }),
  setPrompt: (prompt) => set({ prompt }),
  loadGeneration: (generation) => {
    if (generation.status === "failed") {
      set({
        generationId: generation.id,
        prompt: generation.prompt,
        specification: null,
        files: [],
        activeFilePath: null,
        generationError: generation.error || "Generation failed",
        savedFiles: "[]",
      });
      return;
    }
    if (generation.status !== "completed") {
      return;
    }
    const pages = (generation.specification.pages ?? []).map((page) => ({
      id: page.route === "/" ? "home" : page.route.replace(/^\//, "").replaceAll("/", "-"),
      name: page.name,
      route: page.route,
    }));
    set({
      generationId: generation.id,
      prompt: generation.prompt,
      specification: generation.specification,
      files: generation.files,
      activeFilePath: generation.files[0]?.path ?? null,
      pages: pages.length > 0 ? pages : defaultEditorPages,
      activePageId: pages[0]?.id ?? defaultEditorPages[0].id,
      generationError: null,
      savedFiles: JSON.stringify(generation.files),
    });
  },
  resetWorkspace: () =>
    set({
      files: [],
      activeFilePath: null,
      pages: defaultEditorPages,
      activePageId: defaultEditorPages[0].id,
      prompt: "",
      specification: null,
      generationId: null,
      generationError: null,
      savedFiles: "[]",
    }),
}));
