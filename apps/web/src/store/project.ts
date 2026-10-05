import { create } from "zustand";

export type Project = {
  id: string;
  name: string;
  description?: string;
  hasImage?: boolean;
  latestPrompt?: string;
  latestStatus?: string;
  updatedAt?: string;
};

type ProjectState = {
  currentProject: Project | null;
  projects: Project[];
  setCurrentProject: (project: Project | null) => void;
  setProjects: (projects: Project[]) => void;
};

export const useProjectStore = create<ProjectState>((set) => ({
  currentProject: null,
  projects: [],
  setCurrentProject: (project) => set({ currentProject: project }),
  setProjects: (projects) => set({ projects }),
}));
