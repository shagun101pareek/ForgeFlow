import { create } from "zustand";

type PreviewState = {
  isOpen: boolean;
  url: string | null;
  isLoading: boolean;
  setOpen: (isOpen: boolean) => void;
  setUrl: (url: string | null) => void;
  setLoading: (isLoading: boolean) => void;
};

export const usePreviewStore = create<PreviewState>((set) => ({
  isOpen: false,
  url: null,
  isLoading: false,
  setOpen: (isOpen) => set({ isOpen }),
  setUrl: (url) => set({ url }),
  setLoading: (isLoading) => set({ isLoading }),
}));
