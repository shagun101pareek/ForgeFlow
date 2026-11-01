"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from "@tanstack/react-query";
import axios from "axios";
import { PanelLeft, PanelRight } from "lucide-react";
import dynamic from "next/dynamic";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from "@/components/ui/progress";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { Wordmark } from "@/components/wordmark";
import { examplePrompts } from "@/features/generation/example-prompts";
import {
  GENERATION_PHASES,
  generationPhase,
} from "@/features/generation/generation-phases";
import { AccountMenu } from "@/features/dashboard/account-menu";
import { SourceView } from "@/features/preview/source-view";
import { ProjectImage } from "@/features/project/project-image";
import { PublishDialog, repositoryURL } from "@/features/project/publish-dialog";
import { getErrorMessage } from "@/lib/api-error";
import type { GenerationPromptInput } from "@/lib/validators";
import {
  deleteGeneration,
  downloadGeneration,
  enqueueGeneration,
  generationIsActive,
  getGeneration,
  getLatestGeneration,
  listGenerations,
  saveGeneration,
  waitForGeneration,
  type GenerationSummary,
} from "@/services/generation";
import { getProject, updateProject } from "@/services/projects";
import { useEditorStore, type EditorPage } from "@/store/editor";

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
  const githubUrl = useEditorStore((state) => state.githubUrl);
  const generationError = useEditorStore((state) => state.generationError);
  const activeFilePath = useEditorStore((state) => state.activeFilePath);
  const savedFiles = useEditorStore((state) => state.savedFiles);
  const setActiveFile = useEditorStore((state) => state.setActiveFile);
  const updateFile = useEditorStore((state) => state.updateFile);
  const setActivePage = useEditorStore((state) => state.setActivePage);
  const setPrompt = useEditorStore((state) => state.setPrompt);
  const loadGeneration = useEditorStore((state) => state.loadGeneration);
  const setGithubUrl = useEditorStore((state) => state.setGithubUrl);
  const resetWorkspace = useEditorStore((state) => state.resetWorkspace);
  const activePage = pages.find((page) => page.id === activePageId) ?? pages[0];
  const [draftName, setDraftName] = useState<string | null>(null);
  const [openingVersionId, setOpeningVersionId] = useState<string | null>(null);
  const [downloading, setDownloading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [pendingLoss, setPendingLoss] = useState<
    { kind: "generate" } | { kind: "open"; id: string } | null
  >(null);
  const pendingLossRef = useRef(pendingLoss);
  const unsaved = JSON.stringify(files) !== savedFiles;
  const [canvas, setCanvas] = useState<"preview" | "source">("preview");
  const [publishOpen, setPublishOpen] = useState(false);
  const [navOpen, setNavOpen] = useState(false);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [phaseSnapshot, setPhaseSnapshot] = useState<{
    label: (typeof GENERATION_PHASES)[number];
    index: number;
    progress: number;
  } | null>(null);
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

  function generate(override?: string) {
    const nextPrompt = (override ?? prompt).trim();
    if (!nextPrompt) {
      toast.error("Enter a prompt first");
      return;
    }
    if (unsaved) {
      const next = { kind: "generate" as const };
      pendingLossRef.current = next;
      setPendingLoss(next);
      return;
    }
    generation.mutate({ projectId, prompt: nextPrompt });
  }

  function requestOpenVersion(id: string) {
    if (id === generationId || openingVersionId) {
      return;
    }
    if (unsaved) {
      const next = { kind: "open" as const, id };
      pendingLossRef.current = next;
      setPendingLoss(next);
      return;
    }
    void openVersion(id);
  }

  function confirmLoss() {
    const next = pendingLossRef.current;
    pendingLossRef.current = null;
    setPendingLoss(null);
    if (!next) {
      return;
    }
    if (next.kind === "generate") {
      const nextPrompt = prompt.trim();
      if (!nextPrompt) {
        toast.error("Enter a prompt first");
        return;
      }
      generation.mutate({ projectId, prompt: nextPrompt });
      return;
    }
    void openVersion(next.id);
  }

  async function discardVersion(id: string) {
    await deleteGeneration(projectId, id);
    await queryClient.invalidateQueries({ queryKey: ["generations", projectId] });
    if (id === generationId) {
      const next = await getLatestGeneration(projectId);
      queryClient.setQueryData(["generation-latest", projectId], next);
      loadedProject.current = projectId;
      if (next && !generationIsActive(next.status)) {
        loadGeneration(next);
      } else {
        resetWorkspace();
      }
    }
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

  const showGenerating =
    generation.isPending ||
    (generationIsActive(latest.data?.status) && files.length === 0);

  const activeStatus = generation.isPending
    ? (latest.data?.status ?? "queued")
    : latest.data?.status;

  useEffect(() => {
    if (!showGenerating) {
      return;
    }
    const startedAt = Date.now();
    let timeoutId = 0;
    const tick = () => {
      setPhaseSnapshot(generationPhase(activeStatus, startedAt, Date.now()));
      timeoutId = window.setTimeout(tick, 400);
    };
    timeoutId = window.setTimeout(tick, 0);
    return () => window.clearTimeout(timeoutId);
  }, [showGenerating, activeStatus]);

  useEffect(() => {
    if (!unsaved) {
      return;
    }
    function onBeforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault();
      event.returnValue = "";
    }
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [unsaved]);

  const phase = showGenerating
    ? (phaseSnapshot ?? {
        label: GENERATION_PHASES[0],
        index: 0,
        progress: 8,
      })
    : { label: GENERATION_PHASES[0], index: 0, progress: 0 };

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const meta = event.metaKey || event.ctrlKey;
      if (!meta) {
        return;
      }
      if (event.key === "Enter") {
        if (showGenerating) {
          return;
        }
        event.preventDefault();
        generate();
        return;
      }
      if (event.key.toLowerCase() === "s") {
        if (!generationId || !unsaved || saving || showGenerating) {
          return;
        }
        event.preventDefault();
        void saveEdit();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
    // generate/saveEdit close over the latest editor state each render.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional shortcut binding
  }, [showGenerating, generationId, unsaved, saving, prompt, projectId]);

  const publishDisabled =
    !generationId || files.length === 0 || unsaved || showGenerating;
  const repositoryHref = repositoryURL(githubUrl);

  const activeSections =
    specification?.pages.find((page) => page.route === activePage?.route)
      ?.sections ?? [];

  if (project.isPending) {
    return (
      <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-4 px-4 py-8">
        <Skeleton className="h-9 w-48" />
        <div className="grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)_260px]">
          <div className="space-y-3">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="mt-4 h-4 w-20" />
            <Skeleton className="h-16 w-full" />
          </div>
          <Skeleton className="min-h-[420px] w-full rounded-2xl" />
          <div className="space-y-3">
            <Skeleton className="h-28 w-full rounded-lg" />
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
          </div>
        </div>
      </main>
    );
  }

  if (project.isError || !project.data) {
    return (
      <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 text-center">
        <p className="text-sm text-muted-foreground">
          {getErrorMessage(project.error, "This project could not be opened.")}
        </p>
        <Button
          variant="outline"
          nativeButton={false}
          render={<Link href="/dashboard" />}
        >
          Back to projects
        </Button>
      </main>
    );
  }

  return (
    <div className="flex min-h-dvh flex-col lg:h-dvh">
      <header className="flex flex-wrap items-center gap-3 border-b px-4 py-3">
        <Wordmark href="/dashboard" />
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
          <Sheet open={navOpen} onOpenChange={setNavOpen}>
            <SheetTrigger
              render={
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  className="lg:hidden"
                  aria-label="Pages and versions"
                />
              }
            >
              <PanelLeft className="size-4" />
            </SheetTrigger>
            <SheetContent side="left" className="w-[min(100%,20rem)] overflow-y-auto p-0">
              <SheetHeader>
                <SheetTitle>Pages & versions</SheetTitle>
              </SheetHeader>
              <div className="px-4 pb-6">
                <EditorNavRail
                  pages={pages}
                  activePageId={activePage?.id}
                  onSelectPage={(id) => {
                    setActivePage(id);
                    setNavOpen(false);
                  }}
                  versions={versions}
                  generationId={generationId}
                  openingVersionId={openingVersionId}
                  onOpenVersion={(id) => {
                    setNavOpen(false);
                    requestOpenVersion(id);
                  }}
                  onDiscardVersion={discardVersion}
                />
              </div>
            </SheetContent>
          </Sheet>
          <Sheet open={detailsOpen} onOpenChange={setDetailsOpen}>
            <SheetTrigger
              render={
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  className="lg:hidden"
                  aria-label="Page details"
                />
              }
            >
              <PanelRight className="size-4" />
            </SheetTrigger>
            <SheetContent side="right" className="w-[min(100%,20rem)] overflow-y-auto p-0">
              <SheetHeader>
                <SheetTitle>This page</SheetTitle>
              </SheetHeader>
              <div className="px-4 pb-6">
                <EditorDetailsRail
                  projectId={projectId}
                  hasImage={Boolean(project.data.hasImage)}
                  activePage={activePage}
                  activeSections={activeSections}
                />
              </div>
            </SheetContent>
          </Sheet>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" />}>
              Actions
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                disabled={!unsaved || saving || showGenerating || !generationId}
                onClick={() => void saveEdit()}
              >
                {saving ? "Saving…" : "Save"}
              </DropdownMenuItem>
              <DropdownMenuItem
                disabled={downloading || showGenerating || files.length === 0 || unsaved}
                onClick={() => void download()}
              >
                {downloading ? "Downloading…" : "Download"}
              </DropdownMenuItem>
              <DropdownMenuItem
                disabled={publishDisabled}
                onClick={() => setPublishOpen(true)}
              >
                Publish
              </DropdownMenuItem>
              {repositoryHref ? (
                <DropdownMenuItem
                  nativeButton={false}
                  render={
                    <a href={repositoryHref} target="_blank" rel="noreferrer" />
                  }
                >
                  Open repository
                </DropdownMenuItem>
              ) : null}
            </DropdownMenuContent>
          </DropdownMenu>
          <Button type="button" onClick={() => generate()} disabled={showGenerating}>
            {showGenerating ? phase.label : "Generate"}
          </Button>
          <AccountMenu />
        </div>
      </header>
      <PublishDialog
        projectId={projectId}
        generationId={generationId}
        githubUrl={githubUrl}
        disabled={publishDisabled}
        onPublished={setGithubUrl}
        open={publishOpen}
        onOpenChange={setPublishOpen}
        hideTrigger
      />
      <div className="grid flex-1 grid-cols-1 lg:min-h-0 lg:grid-cols-[220px_minmax(0,1fr)_260px]">
        <aside className="hidden border-b p-4 lg:block lg:overflow-y-auto lg:border-r lg:border-b-0">
          <EditorNavRail
            pages={pages}
            activePageId={activePage?.id}
            onSelectPage={setActivePage}
            versions={versions}
            generationId={generationId}
            openingVersionId={openingVersionId}
            onOpenVersion={requestOpenVersion}
            onDiscardVersion={discardVersion}
          />
        </aside>
        <section className="flex min-h-[420px] flex-col gap-4 bg-muted/40 p-4 lg:min-h-0">
          <div
            id="forgeflow-preview"
            className="flex min-h-[min(70vh,560px)] flex-1 flex-col overflow-hidden rounded-2xl border bg-card shadow-lg lg:min-h-[560px]"
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
              <div className="flex flex-1 flex-col items-center justify-center gap-6 px-6 text-center">
                {openingVersionId ? (
                  <>
                    <p className="text-base font-medium">Opening this version</p>
                    <p className="max-w-sm text-sm text-muted-foreground">
                      The preview will replace what is on the canvas.
                    </p>
                  </>
                ) : (
                  <>
                    <div className="w-full max-w-md">
                      <Progress value={phase.progress} className="w-full">
                        <ProgressLabel>{phase.label}</ProgressLabel>
                        <ProgressValue />
                      </Progress>
                    </div>
                    <ol className="flex w-full max-w-md flex-col gap-2 text-left sm:flex-row sm:justify-between">
                      {GENERATION_PHASES.map((label, index) => {
                        const done = index < phase.index;
                        const current = index === phase.index;
                        return (
                          <li
                            key={label}
                            className={`text-xs transition-colors duration-300 ${
                              current
                                ? "font-medium text-foreground"
                                : done
                                  ? "text-foreground/70"
                                  : "text-muted-foreground"
                            }`}
                          >
                            <span
                              className={`mr-1.5 inline-block h-1.5 w-1.5 rounded-full align-middle transition-colors duration-300 ${
                                current || done ? "bg-primary" : "bg-muted-foreground/40"
                              }`}
                            />
                            {label}
                          </li>
                        );
                      })}
                    </ol>
                    <p className="max-w-sm text-sm text-muted-foreground">
                      The preview opens here when it is ready.
                    </p>
                  </>
                )}
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
              <div className="flex flex-1 flex-col justify-center gap-5 px-6 py-10">
                {generationError ? (
                  <div className="max-w-md">
                    <p className="font-heading text-2xl font-medium">
                      Generation failed
                    </p>
                    <p className="mt-2 text-sm text-muted-foreground">
                      {generationError}
                    </p>
                    <div className="mt-4 flex flex-wrap gap-2">
                      <Button
                        type="button"
                        onClick={() => generate()}
                        disabled={!prompt.trim()}
                      >
                        Try again
                      </Button>
                    </div>
                    <p className="mt-6 text-sm text-muted-foreground">
                      Or start from a layout below.
                    </p>
                  </div>
                ) : (
                  <div>
                    <p className="font-heading text-3xl font-medium">
                      Start from a layout
                    </p>
                    <p className="mt-2 max-w-md text-sm text-muted-foreground">
                      Pick one and the preview starts. You can still edit the
                      prompt below.
                    </p>
                  </div>
                )}
                <div className="grid gap-3 sm:grid-cols-2">
                  {examplePrompts.map((example) => (
                    <button
                      key={example.id}
                      type="button"
                      onClick={() => {
                        setPrompt(example.text);
                        generate(example.text);
                      }}
                      className={`rounded-2xl p-4 text-left ${example.tone}`}
                    >
                      <span className="font-heading text-xl">{example.label}</span>
                      <span className="mt-2 block text-sm leading-6 opacity-80">
                        {example.text}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            ) : null}
          </div>
          <div className="rounded-xl border bg-card p-3">
            <div className="mb-3 flex flex-wrap gap-2">
              {examplePrompts.map((example) => (
                <button
                  key={example.id}
                  type="button"
                  onClick={() => setPrompt(example.text)}
                  className={`rounded-full px-2.5 py-1 text-xs ${example.tone} ${
                    prompt === example.text ? "ring-2 ring-foreground" : ""
                  }`}
                >
                  {example.label}
                </button>
              ))}
            </div>
            <Textarea
              aria-label="Prompt"
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Describe the interface you want to generate."
              className="min-h-24"
            />
            <div className="mt-3 flex justify-end">
              <Button type="button" onClick={() => generate()} disabled={showGenerating}>
                {showGenerating ? phase.label : "Generate"}
              </Button>
            </div>
          </div>
        </section>
        <aside className="hidden border-t p-4 lg:block lg:overflow-y-auto lg:border-t-0 lg:border-l">
          <EditorDetailsRail
            projectId={projectId}
            hasImage={Boolean(project.data.hasImage)}
            activePage={activePage}
            activeSections={activeSections}
          />
        </aside>
      </div>
      <AlertDialog
        open={pendingLoss !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingLoss(null);
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Discard unsaved edits?</AlertDialogTitle>
            <AlertDialogDescription>
              Source changes on this version are not saved. Continuing drops them.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Keep editing</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={confirmLoss}>
              Continue
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function EditorNavRail({
  pages,
  activePageId,
  onSelectPage,
  versions,
  generationId,
  openingVersionId,
  onOpenVersion,
  onDiscardVersion,
}: {
  pages: EditorPage[];
  activePageId?: string;
  onSelectPage: (id: string) => void;
  versions: UseQueryResult<GenerationSummary[]>;
  generationId: string | null;
  openingVersionId: string | null;
  onOpenVersion: (id: string) => void;
  onDiscardVersion: (id: string) => Promise<void>;
}) {
  return (
    <>
      <p className="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
        Pages
      </p>
      <ul className="flex flex-col gap-2">
        {pages.map((page) => {
          const selected = page.id === activePageId;
          return (
            <li key={page.id}>
              <button
                type="button"
                onClick={() => onSelectPage(page.id)}
                className={`w-full rounded-lg px-3 py-2 text-left text-sm ${
                  selected ? "bg-muted font-medium" : "hover:bg-muted/60"
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
        <div className="space-y-2" aria-busy="true">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-12 w-full" />
        </div>
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
        <ul className="flex flex-col gap-2">
          {versions.data.map((version) => (
            <VersionItem
              key={version.id}
              version={version}
              selected={version.id === generationId}
              opening={openingVersionId === version.id}
              onOpen={() => onOpenVersion(version.id)}
              onDiscard={() => onDiscardVersion(version.id)}
            />
          ))}
        </ul>
      ) : null}
    </>
  );
}

function EditorDetailsRail({
  projectId,
  hasImage,
  activePage,
  activeSections,
}: {
  projectId: string;
  hasImage: boolean;
  activePage?: EditorPage;
  activeSections: { type: string; title?: string }[];
}) {
  return (
    <>
      <ProjectImage projectId={projectId} hasImage={hasImage} />
      <p className="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
        This page
      </p>
      {activePage ? (
        <div className="text-sm">
          <p className="font-medium">{activePage.name}</p>
          <p className="mt-1 text-muted-foreground">{activePage.route}</p>
          {activeSections.length > 0 ? (
            <ul className="mt-4 space-y-2">
              {activeSections.map((section, index) => (
                <li key={`${section.type}-${index}`}>
                  <span className="capitalize">{section.type}</span>
                  {section.title ? (
                    <span className="mt-0.5 block text-muted-foreground">
                      {section.title}
                    </span>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : (
            <p className="mt-4 text-muted-foreground">No sections yet.</p>
          )}
        </div>
      ) : null}
    </>
  );
}

function versionStatusLabel(status: string) {
  if (status === "completed" || status === "failed") {
    return status === "failed" ? "Failed · " : "";
  }
  if (status === "queued") {
    return "Understanding prompt · ";
  }
  if (status === "running") {
    return "Writing React · ";
  }
  return `${status.charAt(0).toUpperCase()}${status.slice(1)} · `;
}

function VersionItem({
  version,
  selected,
  opening,
  onOpen,
  onDiscard,
}: {
  version: GenerationSummary;
  selected: boolean;
  opening: boolean;
  onOpen: () => void;
  onDiscard: () => Promise<void>;
}) {
  const [open, setOpen] = useState(false);
  const [removing, setRemoving] = useState(false);
  const label =
    version.prompt.length > 72 ? `${version.prompt.slice(0, 72)}…` : version.prompt;
  const running = generationIsActive(version.status);

  async function remove() {
    if (removing) {
      return;
    }
    setRemoving(true);
    try {
      await onDiscard();
      toast.success("Version removed");
      setOpen(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not remove that version"));
    } finally {
      setRemoving(false);
    }
  }

  return (
    <li className="flex min-w-56 items-start gap-1 lg:min-w-0">
      <button
        type="button"
        onClick={onOpen}
        disabled={opening}
        className={`min-w-0 flex-1 rounded-lg px-3 py-2 text-left text-sm ${
          selected ? "bg-muted font-medium" : "hover:bg-muted/60"
        }`}
      >
        {label}
        <span className="mt-0.5 block text-xs text-muted-foreground">
          {versionStatusLabel(version.status)}
          {new Date(version.createdAt).toLocaleString()}
        </span>
      </button>
      {selected && !running ? (
        <AlertDialog open={open} onOpenChange={setOpen}>
          <AlertDialogTrigger
            render={<Button variant="ghost" size="sm" className="text-muted-foreground" />}
          >
            Discard
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Discard this version?</AlertDialogTitle>
              <AlertDialogDescription>
                This removes the version from the project. Other versions stay,
                and a published GitHub repository is left in place.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                variant="destructive"
                disabled={removing}
                onClick={() => void remove()}
              >
                {removing ? "Removing…" : "Discard"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      ) : null}
    </li>
  );
}
