"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import dynamic from "next/dynamic";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { getErrorMessage } from "@/lib/api-error";
import type { GenerationPromptInput } from "@/lib/validators";
import {
  downloadGeneration,
  enqueueGeneration,
  generationIsActive,
  getGeneration,
  getLatestGeneration,
  listGenerations,
  saveGeneration,
  waitForGeneration,
} from "@/services/generation";
import { PublishDialog } from "@/features/project/publish-dialog";
import { SourceView } from "@/features/preview/source-view";
import { getProject, updateProject } from "@/services/projects";
import { useEditorStore } from "@/store/editor";

const LivePreview = dynamic(
  () =>
    import("@/features/preview/live-preview").then((mod) => mod.LivePreview),
  { ssr: false },
);

export function ProjectEditor({ projectId }: { projectId: string }) {
  const queryClient = useQueryClient();
  const project = useQuery({
    queryKey: ["project", projectId],
    queryFn: () => getProject(projectId),
    retry: false,
  });
  const pages = useEditorStore((state) => state.pages);
  const activePageId = useEditorStore((state) => state.activePageId);
  const prompt = useEditorStore((state) => state.prompt);
  const files = useEditorStore((state) => state.files);
  const specification = useEditorStore((state) => state.specification);
  const generationId = useEditorStore((state) => state.generationId);
  const generationError = useEditorStore((state) => state.generationError);
  const activeFilePath = useEditorStore((state) => state.activeFilePath);
  const savedFiles = useEditorStore((state) => state.savedFiles);
  const setActiveFile = useEditorStore((state) => state.setActiveFile);
  const updateFile = useEditorStore((state) => state.updateFile);
  const setActivePage = useEditorStore((state) => state.setActivePage);
  const setPrompt = useEditorStore((state) => state.setPrompt);
  const loadGeneration = useEditorStore((state) => state.loadGeneration);
  const resetWorkspace = useEditorStore((state) => state.resetWorkspace);
  const activePage = pages.find((page) => page.id === activePageId) ?? pages[0];
  const [draftName, setDraftName] = useState<string | null>(null);
  const [openingVersionId, setOpeningVersionId] = useState<string | null>(null);
  const [downloading, setDownloading] = useState(false);
  const [saving, setSaving] = useState(false);
  const unsaved = JSON.stringify(files) !== savedFiles;
  const [canvas, setCanvas] = useState<"preview" | "source">("preview");
  const name = draftName ?? project.data?.name ?? "";
  const loadedProject = useRef<string | null>(null);

  const versions = useQuery({
    queryKey: ["generations", projectId],
    queryFn: () => listGenerations(projectId),
    refetchInterval: (query) =>
      query.state.data?.some((version) => generationIsActive(version.status))
        ? 500
        : false,
  });
  const latest = useQuery({
    queryKey: ["generation-latest", projectId],
    queryFn: () => getLatestGeneration(projectId),
    retry: false,
    refetchInterval: (query) =>
      generationIsActive(query.state.data?.status) ? 500 : false,
  });

  useEffect(() => {
    loadedProject.current = null;
    resetWorkspace();
  }, [projectId, resetWorkspace]);

  useEffect(() => {
    if (latest.isPending || latest.isFetching) {
      return;
    }
    if (loadedProject.current === projectId) {
      return;
    }
    if (!latest.data || generationIsActive(latest.data.status)) {
      return;
    }
    loadedProject.current = projectId;
    loadGeneration(latest.data);
  }, [
    latest.data,
    latest.isFetching,
    latest.isPending,
    loadGeneration,
    projectId,
  ]);

  const rename = useMutation({
    mutationFn: (nextName: string) =>
      updateProject(projectId, { name: nextName }),
    onSuccess: async (updated) => {
      setDraftName(null);
      queryClient.setQueryData(["project", projectId], updated);
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
    onError: (error) => {
      setDraftName(null);
      toast.error(getErrorMessage(error, "Could not rename project"));
    },
  });

  function saveName() {
    const nextName = name.trim();
    if (!project.data || !nextName || nextName === project.data.name) {
      setDraftName(null);
      return;
    }
    rename.mutate(nextName);
  }

  const generation = useMutation({
    mutationFn: async (input: GenerationPromptInput) => {
      const queued = await enqueueGeneration(input);
      await queryClient.invalidateQueries({ queryKey: ["generations", projectId] });
      return waitForGeneration(input.projectId, queued.id);
    },
    onSuccess: (result) => {
      loadedProject.current = projectId;
      loadGeneration(result);
      queryClient.setQueryData(["generation-latest", projectId], result);
      void queryClient.invalidateQueries({ queryKey: ["generations", projectId] });
      toast.success("Preview is ready");
    },
    onError: (error) => {
      void queryClient.invalidateQueries({ queryKey: ["generations", projectId] });
      const message =
        error instanceof Error && !axios.isAxiosError(error) && error.message
          ? error.message
          : getErrorMessage(error, "Could not generate the interface");
      toast.error(message);
    },
  });

  function generate() {
    const nextPrompt = prompt.trim();
    if (!nextPrompt) {
      toast.error("Enter a prompt first");
      return;
    }
    generation.mutate({ projectId, prompt: nextPrompt });
  }

  async function openVersion(id: string) {
    if (id === generationId || openingVersionId) {
      return;
    }
    setOpeningVersionId(id);
    try {
      const record = await getGeneration(projectId, id);
      if (generationIsActive(record.status)) {
        return;
      }
      loadedProject.current = projectId;
      loadGeneration(record);
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not open that version"));
    } finally {
      setOpeningVersionId(null);
    }
  }

  async function saveEdit() {
    if (!generationId || !unsaved || saving) {
      return;
    }
    setSaving(true);
    try {
      const result = await saveGeneration(projectId, generationId, files);
      loadedProject.current = projectId;
      loadGeneration(result);
      queryClient.setQueryData(["generation-latest", projectId], result);
      await queryClient.invalidateQueries({ queryKey: ["generations", projectId] });
      toast.success("Edit saved as a new version");
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not save this edit"));
    } finally {
      setSaving(false);
    }
  }

  async function download() {
    if (!generationId || files.length === 0 || downloading || unsaved) {
      toast(unsaved ? "Save the edit before downloading" : "Generate a UI first");
      return;
    }
    setDownloading(true);
    try {
      await downloadGeneration(projectId, generationId);
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not download this generation"));
    } finally {
      setDownloading(false);
    }
  }

  function showPreview() {
    if (files.length === 0) {
      toast("Generate a UI first");
      return;
    }
    setCanvas("preview");
    document.getElementById("forgeflow-preview")?.scrollIntoView({
      behavior: "smooth",
      block: "nearest",
    });
  }

  const showGenerating =
    generation.isPending ||
    (generationIsActive(latest.data?.status) && files.length === 0);

  const activeSections =
    specification?.pages.find((page) => page.route === activePage?.route)
      ?.sections ?? [];

  if (project.isPending) {
    return (
      <main className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
        Loading project…
      </main>
    );
  }

  if (project.isError || !project.data) {
    return (
      <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 text-center">
        <p className="text-sm text-muted-foreground">
          {getErrorMessage(project.error, "This project could not be opened.")}
        </p>
        <Button variant="outline" render={<Link href="/dashboard" />}>
          Back to projects
        </Button>
      </main>
    );
  }

  return (
    <div className="flex min-h-dvh flex-col lg:h-dvh">
      <header className="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <Link
          href="/dashboard"
          className="text-sm font-semibold tracking-tight"
        >
          ForgeFlow
        </Link>
        <form
          className="min-w-0 flex-1"
          onSubmit={(event) => {
            event.preventDefault();
            saveName();
          }}
        >
          <Input
            aria-label="Project name"
            value={name}
            onChange={(event) => setDraftName(event.target.value)}
            onBlur={saveName}
            disabled={rename.isPending}
            className="max-w-sm font-medium"
          />
        </form>
        <div className="flex items-center gap-2">
          <Button variant="outline" type="button" onClick={showPreview}>
            Preview
          </Button>
          <PublishDialog
            projectId={projectId}
            generationId={generationId}
            disabled={!generationId || files.length === 0 || unsaved || showGenerating}
          />
          <Button
            variant="outline"
            type="button"
            onClick={() => void saveEdit()}
            disabled={!unsaved || saving || showGenerating || !generationId}
          >
            {saving ? "Saving…" : "Save"}
          </Button>
          <Button
            variant="outline"
            type="button"
            onClick={() => void download()}
            disabled={downloading || showGenerating || files.length === 0 || unsaved}
          >
            {downloading ? "Downloading…" : "Download"}
          </Button>
          <Button type="button" onClick={generate} disabled={showGenerating}>
            {showGenerating ? "Generating…" : "Generate"}
          </Button>
        </div>
      </header>
      <div className="grid flex-1 grid-cols-1 lg:min-h-0 lg:grid-cols-[220px_minmax(0,1fr)_260px]">
        <aside className="border-b p-4 lg:overflow-y-auto lg:border-r lg:border-b-0">
          <p className="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
            Pages
          </p>
          <ul className="flex gap-2 overflow-x-auto lg:flex-col">
            {pages.map((page) => {
              const selected = page.id === activePage?.id;
              return (
                <li key={page.id}>
                  <button
                    type="button"
                    onClick={() => setActivePage(page.id)}
                    className={`w-full rounded-lg px-3 py-2 text-left text-sm ${
                      selected
                        ? "bg-muted font-medium"
                        : "hover:bg-muted/60"
                    }`}
                  >
                    {page.name}
                    <span className="mt-0.5 block text-xs text-muted-foreground">
                      {page.route}
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
          <p className="mt-6 mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
            Versions
          </p>
          {versions.isPending ? (
            <p className="text-sm text-muted-foreground">Loading versions…</p>
          ) : null}
          {versions.isError ? (
            <p className="text-sm text-muted-foreground">
              Versions could not be loaded.
            </p>
          ) : null}
          {versions.data && versions.data.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Generated designs will be saved here.
            </p>
          ) : null}
          {versions.data && versions.data.length > 0 ? (
            <ul className="flex gap-2 overflow-x-auto lg:flex-col">
              {versions.data.map((version) => {
                const selected = version.id === generationId;
                const label =
                  version.prompt.length > 72
                    ? `${version.prompt.slice(0, 72)}…`
                    : version.prompt;
                return (
                  <li key={version.id}>
                    <button
                      type="button"
                      onClick={() => void openVersion(version.id)}
                      disabled={openingVersionId === version.id}
                      className={`w-full rounded-lg px-3 py-2 text-left text-sm ${
                        selected ? "bg-muted font-medium" : "hover:bg-muted/60"
                      }`}
                    >
                      {label}
                      <span className="mt-0.5 block text-xs text-muted-foreground">
                        {versionStatusLabel(version.status)}
                        {new Date(version.createdAt).toLocaleString()}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          ) : null}
        </aside>
        <section className="flex min-h-[420px] flex-col gap-4 p-4 lg:min-h-0">
          <div
            id="forgeflow-preview"
            className="flex min-h-[560px] flex-1 flex-col overflow-hidden rounded-xl border"
          >
            {!showGenerating && !openingVersionId && files.length > 0 ? (
              <div className="flex gap-1 border-b p-2">
                <button
                  type="button"
                  onClick={() => setCanvas("preview")}
                  className={`rounded-md px-2.5 py-1 text-xs ${
                    canvas === "preview" ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
                  }`}
                >
                  Preview
                </button>
                <button
                  type="button"
                  onClick={() => setCanvas("source")}
                  className={`rounded-md px-2.5 py-1 text-xs ${
                    canvas === "source" ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
                  }`}
                >
                  Source
                </button>
              </div>
            ) : null}
            {showGenerating || openingVersionId ? (
              <div className="flex flex-1 items-center justify-center px-6 text-center text-sm text-muted-foreground">
                {openingVersionId ? "Opening version…" : "Generating the interface…"}
              </div>
            ) : null}
            {!showGenerating && !openingVersionId && files.length > 0 && canvas === "preview" ? (
              <LivePreview
                key={`${generationId ?? "draft"}:${activePage?.route ?? "/"}:${files.reduce((total, file) => total + file.code.length, 0)}`}
                files={files}
                route={activePage?.route ?? "/"}
              />
            ) : null}
            {!showGenerating && !openingVersionId && files.length > 0 && canvas === "source" ? (
              <SourceView
                files={files}
                activePath={activeFilePath}
                onSelect={setActiveFile}
                onChange={updateFile}
              />
            ) : null}
            {!showGenerating && !openingVersionId && files.length === 0 ? (
              <div className="flex flex-1 flex-col items-center justify-center px-6 py-16 text-center">
                <p className="text-sm font-medium">
                  {generationError ? "Generation failed" : activePage?.name}
                </p>
                <p className="mt-2 max-w-md text-sm text-muted-foreground">
                  {generationError ??
                    "Describe an interface below, then generate it. The preview will run here."}
                </p>
              </div>
            ) : null}
          </div>
          <label className="flex flex-col gap-2 text-sm font-medium">
            Prompt
            <Textarea
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Describe the interface you want to generate."
              className="min-h-24"
            />
          </label>
        </section>
        <aside className="border-t p-4 lg:overflow-y-auto lg:border-t-0 lg:border-l">
          <p className="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
            Inspector
          </p>
          <p className="text-sm text-muted-foreground">
            {activeSections.length > 0
              ? "Sections on this page."
              : "Component details will show up here after a generation."}
          </p>
          {activePage ? (
            <dl className="mt-6 space-y-3 text-sm">
              <div>
                <dt className="text-muted-foreground">Page</dt>
                <dd>{activePage.name}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Route</dt>
                <dd>{activePage.route}</dd>
              </div>
              {activeSections.length > 0 ? (
                <div>
                  <dt className="text-muted-foreground">Sections</dt>
                  <dd>
                    <ul className="mt-1 space-y-1">
                      {activeSections.map((section, index) => (
                        <li key={`${section.type}-${index}`}>
                          {section.type}
                          {section.title ? ` · ${section.title}` : ""}
                        </li>
                      ))}
                    </ul>
                  </dd>
                </div>
              ) : null}
            </dl>
          ) : null}
        </aside>
      </div>
    </div>
  );
}

function versionStatusLabel(status: string) {
  if (status === "completed") {
    return "";
  }
  return `${status.charAt(0).toUpperCase()}${status.slice(1)} · `;
}
