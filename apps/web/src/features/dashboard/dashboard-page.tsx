"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { MoreHorizontal } from "lucide-react";
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
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Wordmark } from "@/components/wordmark";
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
        <Wordmark href="/dashboard" />
        <div className="flex items-center gap-2">
          <CreateProjectDialog />
          <AccountMenu />
        </div>
      </header>
      <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 px-4 py-8 sm:px-6">
        <div>
          <h1 className="font-heading text-4xl font-medium tracking-tight">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Open a project to enter the editor.
          </p>
        </div>
        {projects.isPending ? (
          <p className="text-sm text-muted-foreground">Loading projects…</p>
        ) : null}
        {projects.isError ? (
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-sm text-destructive">
              {getErrorMessage(projects.error, "Could not load projects")}
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => void projects.refetch()}
            >
              Retry
            </Button>
          </div>
        ) : null}
        {projects.data && projects.data.length === 0 ? (
          <div className="max-w-lg py-10">
            <p className="font-heading text-3xl font-medium">No projects yet</p>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Create one and the editor will offer a landing page, a dashboard, a
              portfolio, or a pricing page to generate.
            </p>
            <div className="mt-6">
              <CreateProjectDialog />
            </div>
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

const projectTones = [
  "bg-primary",
  "bg-teal-600",
  "bg-rose-500",
  "bg-amber-600",
];

function projectTone(name: string) {
  const total = [...name].reduce((sum, char) => sum + char.charCodeAt(0), 0);
  return projectTones[total % projectTones.length];
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

function StatusBadge({ status }: { status: string | undefined }) {
  if (status === "queued" || status === "running") {
    return <Badge variant="secondary">Running</Badge>;
  }
  if (status === "failed") {
    return <Badge variant="destructive">Failed</Badge>;
  }
  if (status === "completed") {
    return <Badge variant="outline">Ready</Badge>;
  }
  return null;
}

function relativeTime(value: string) {
  const seconds = Math.round((Date.now() - new Date(value).getTime()) / 1000);
  if (seconds < 60) {
    return "just now";
  }
  const ranges: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ["year", 60 * 60 * 24 * 365],
    ["month", 60 * 60 * 24 * 30],
    ["day", 60 * 60 * 24],
    ["hour", 60 * 60],
    ["minute", 60],
  ];
  const format = new Intl.RelativeTimeFormat("en", { numeric: "auto" });
  for (const [unit, size] of ranges) {
    if (seconds >= size) {
      return format.format(-Math.round(seconds / size), unit);
    }
  }
  return "just now";
}

function ProjectCard({ project }: { project: Project }) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const updated = project.updatedAt ? relativeTime(project.updatedAt) : null;

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
              <span
                aria-hidden="true"
                className={`mb-3 block h-1.5 w-10 rounded-full ${projectTone(project.name)}`}
              />
              <div className="flex flex-wrap items-center gap-2">
                <CardTitle className="truncate">{project.name}</CardTitle>
                <StatusBadge status={project.latestStatus} />
              </div>
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
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Actions for ${project.name}`}
                />
              }
            >
              <MoreHorizontal />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                disabled={duplicate.isPending || mutation.isPending}
                onClick={() => duplicate.mutate()}
              >
                {duplicate.isPending ? "Copying…" : "Duplicate"}
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                disabled={duplicate.isPending || mutation.isPending}
                onClick={() => setOpen(true)}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <AlertDialog open={open} onOpenChange={setOpen}>
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
      </CardHeader>
    </Card>
  );
}
