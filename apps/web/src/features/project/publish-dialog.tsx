"use client";

import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getErrorMessage } from "@/lib/api-error";
import { readGitHubToken, writeGitHubToken } from "@/lib/github-token";
import { publishGeneration } from "@/services/generation";

export function repositoryURL(value: string) {
  try {
    const parsed = new URL(value);
    if (
      parsed.protocol !== "https:" ||
      parsed.hostname !== "github.com" ||
      parsed.username ||
      parsed.password ||
      parsed.search ||
      parsed.hash
    ) {
      return null;
    }
    const parts = parsed.pathname.split("/").filter(Boolean);
    if (parts.length !== 2) {
      return null;
    }
    return parsed.toString().replace(/\/$/, "");
  } catch {
    return null;
  }
}

export function PublishDialog({
  projectId,
  generationId,
  githubUrl,
  disabled,
  onPublished,
  open: openProp,
  onOpenChange,
  hideTrigger = false,
}: {
  projectId: string;
  generationId: string | null;
  githubUrl: string;
  disabled: boolean;
  onPublished: (url: string) => void;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  hideTrigger?: boolean;
}) {
  const [localOpen, setLocalOpen] = useState(false);
  const open = openProp ?? localOpen;
  const [token, setToken] = useState("");
  const [rememberToken, setRememberToken] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [publishedUrl, setPublishedUrl] = useState<string | null>(null);
  const savedUrl = repositoryURL(githubUrl);
  const url = publishedUrl ?? savedUrl;

  function close(nextOpen: boolean) {
    if (nextOpen) {
      const stored = readGitHubToken();
      setToken(stored);
      if (stored) {
        setRememberToken(true);
      }
    } else {
      setPublishedUrl(null);
      setToken(readGitHubToken());
    }
    if (onOpenChange) {
      onOpenChange(nextOpen);
    } else {
      setLocalOpen(nextOpen);
    }
  }

  async function publish() {
    if (!generationId || publishing) {
      return;
    }
    const nextToken = token.trim();
    if (nextToken.length < 20) {
      toast.error("Paste a GitHub personal access token with the repo scope");
      return;
    }
    setPublishing(true);
    try {
      const result = await publishGeneration(projectId, generationId, nextToken);
      const nextUrl = repositoryURL(result.url);
      if (!nextUrl) {
        toast.error("GitHub did not return a repository url");
        return;
      }
      if (rememberToken) {
        writeGitHubToken(nextToken);
      } else {
        writeGitHubToken("");
        setToken("");
      }
      setPublishedUrl(nextUrl);
      onPublished(nextUrl);
      toast.success("Repository created");
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not publish to GitHub"));
    } finally {
      setPublishing(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={close}>
      {hideTrigger ? null : (
        <DialogTrigger
          render={
            <Button variant="outline" type="button" disabled={disabled} />
          }
        >
          GitHub
        </DialogTrigger>
      )}
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Publish to GitHub</DialogTitle>
          <DialogDescription>
            Creates a private repository from this version. Use a classic
            personal access token with the{" "}
            <span className="font-medium">repo</span> scope. Manage a saved
            token in{" "}
            <Link href="/settings" className="underline">
              Settings
            </Link>
            .
          </DialogDescription>
        </DialogHeader>
        {url ? (
          <a
            href={url}
            target="_blank"
            rel="noreferrer"
            className="text-sm font-medium underline"
          >
            {url}
          </a>
        ) : (
          <p className="text-sm text-muted-foreground">
            No repository yet. Publish this version to create one under your
            GitHub account.
          </p>
        )}
        {publishedUrl ? null : (
          <form
            className="flex flex-col gap-4"
            onSubmit={(event) => {
              event.preventDefault();
              void publish();
            }}
          >
            <div className="flex flex-col gap-2">
              <Label htmlFor="github-token">Personal access token</Label>
              <Input
                id="github-token"
                type="password"
                autoComplete="off"
                placeholder="ghp_…"
                value={token}
                onChange={(event) => setToken(event.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Create one with the repo scope at{" "}
                <a
                  href="https://github.com/settings/tokens/new?scopes=repo&description=ForgeFlow"
                  target="_blank"
                  rel="noreferrer"
                  className="underline"
                >
                  GitHub token settings
                </a>
                . Fine-grained tokens need repository create permission.
              </p>
              <label className="flex items-center gap-2 text-xs text-muted-foreground">
                <input
                  type="checkbox"
                  className="size-3.5 accent-foreground"
                  checked={rememberToken}
                  onChange={(event) => setRememberToken(event.target.checked)}
                />
                Remember this token on this device
              </label>
            </div>
            <DialogFooter>
              <Button type="submit" disabled={publishing || token.trim().length < 20}>
                {publishing ? "Publishing…" : "Publish"}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
