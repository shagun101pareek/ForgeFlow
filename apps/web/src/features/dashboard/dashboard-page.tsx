"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
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
import { Input } from "@/components/ui/input";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { getErrorMessage } from "@/lib/api-error";
import { repositoryURL } from "@/features/project/publish-dialog";
import { deleteProject, duplicateProject, listProjects } from "@/services/projects";
import type { Project } from "@/store/project";

import { AccountMenu } from "./account-menu";
import { CreateProjectDialog } from "./create-project-dialog";

export function DashboardPage() {
  const [query, setQuery] = useState("");
  const projects = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
    refetchInterval: (projectsQuery) =>
      projectsQuery.state.data?.some((project) =>
        project.latestStatus === "queued" || project.latestStatus === "running",
      )
        ? 1000
        : false,
  });
  const needle = query.trim().toLowerCase();
  const visible = (projects.data ?? []).filter((project) =>
    projectMatches(project, needle),
  );

  return (
    <div className="flex flex-1 flex-col">
      <header className="flex items-center justify-between gap-4 border-b px-4 py-3 sm:px-6">
        <Link href="/dashboard" className="text-sm font-semibold tracking-tight">
          ForgeFlow
        </Link>
        <div className="flex items-center gap-2">
          <CreateProjectDialog />
          <AccountMenu />
        </div>
      </header>
      <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 px-4 py-8 sm:px-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Open a project to enter the editor.
          </p>
        </div>
        {projects.isPending ? (
          <p className="text-sm text-muted-foreground">Loading projects…</p>
        ) : null}
        {projects.isError ? (
          <p className="text-sm text-destructive">
            {getErrorMessage(projects.error, "Could not load projects")}
          </p>
        ) : null}
        {projects.data && projects.data.length === 0 ? (
          <div className="rounded-xl border border-dashed px-6 py-16 text-center">
            <p className="text-sm text-muted-foreground">
              No projects yet. Create one to start a workspace.
            </p>
          </div>
        ) : null}
        {projects.data && projects.data.length > 0 ? (
          <Input
            aria-label="Find a project"
            placeholder="Find by name or prompt"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="max-w-sm"
          />
        ) : null}
        {projects.data && projects.data.length > 0 && visible.length === 0 ? (
          <p className="text-sm text-muted-foreground">No projects match that search.</p>
        ) : null}
        {visible.length > 0 ? (
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {visible.map((project) => (
              <li key={project.id}>
                <ProjectCard project={project} />
              </li>
            ))}
          </ul>
        ) : null}
      </main>
    </div>
  );
}

function projectMatches(project: Project, needle: string) {
  if (!needle) {
    return true;
  }
  return (
    project.name.toLowerCase().includes(needle) ||
    (project.latestPrompt ?? "").toLowerCase().includes(needle)
  );
}

function RepositoryLink({ url }: { url: string | undefined }) {
  const href = url ? repositoryURL(url) : null;
  if (!href) {
    return null;
  }
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="mt-2 block truncate text-xs underline"
    >
      {href.replace("https://github.com/", "")}
    </a>
  );
}

function OutcomeLabel({ status }: { status: string | undefined }) {
  if (status === "queued" || status === "running") {
    return <span className="font-medium">Running. </span>;
  }
  if (status === "failed") {
    return <span className="font-medium text-destructive">Failed. </span>;
  }
  return null;
}

function ProjectCard({ project }: { project: Project }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const updated = project.updatedAt
    ? new Date(project.updatedAt).toLocaleString()
    : null;

  const duplicate = useMutation({
    mutationFn: () => duplicateProject(project.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Project duplicated");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Could not duplicate project"));
    },
  });

  const mutation = useMutation({
    mutationFn: () => deleteProject(project.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Project deleted");
      setOpen(false);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Could not delete project"));
    },
  });

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <Link href={`/projects/${project.id}`} className="block">
              <CardTitle className="truncate">{project.name}</CardTitle>
              {project.description ? (
                <CardDescription className="mt-1 line-clamp-2">
                  {project.description}
                </CardDescription>
              ) : null}
              <p
                className={`mt-3 line-clamp-2 text-sm ${
                  project.latestPrompt ? "text-foreground" : "text-muted-foreground"
                }`}
              >
                <OutcomeLabel status={project.latestStatus} />
                {project.latestPrompt ?? "Nothing generated yet."}
              </p>
              {updated ? (
                <p className="mt-3 text-xs text-muted-foreground">
                  Updated {updated}
                </p>
              ) : null}
            </Link>
            <RepositoryLink url={project.latestRepository} />
          </div>
          <div className="flex flex-col items-end gap-1">
            <Button
              variant="ghost"
              size="sm"
              type="button"
              onClick={() => duplicate.mutate()}
              disabled={duplicate.isPending || mutation.isPending}
            >
              {duplicate.isPending ? "Copying…" : "Duplicate"}
            </Button>
            <AlertDialog open={open} onOpenChange={setOpen}>
            <AlertDialogTrigger
              render={<Button variant="ghost" size="sm" disabled={duplicate.isPending} />}
            >
              Delete
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete {project.name}?</AlertDialogTitle>
                <AlertDialogDescription>
                  This removes the project from your account.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  disabled={mutation.isPending}
                  onClick={() => mutation.mutate()}
                >
                  {mutation.isPending ? "Deleting…" : "Delete"}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </CardHeader>
    </Card>
  );
}
