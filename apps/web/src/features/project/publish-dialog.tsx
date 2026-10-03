"use client";

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
import { publishGeneration } from "@/services/generation";

function repositoryURL(value: string) {
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
}: {
  projectId: string;
  generationId: string | null;
  githubUrl: string;
  disabled: boolean;
  onPublished: (url: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [publishedUrl, setPublishedUrl] = useState<string | null>(null);
  const savedUrl = repositoryURL(githubUrl);
  const url = publishedUrl ?? savedUrl;

  function close(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setToken("");
      setPublishedUrl(null);
    }
  }

  async function publish() {
    if (!generationId || publishing) {
      return;
    }
    const nextToken = token.trim();
    if (nextToken.length < 20) {
      toast.error("Paste a GitHub personal access token");
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
      setToken("");
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
    <div className="flex items-center gap-2">
      {savedUrl ? (
        <a
          href={savedUrl}
          target="_blank"
          rel="noreferrer"
          className="max-w-40 truncate text-sm underline"
        >
          {savedUrl.replace("https://github.com/", "")}
        </a>
      ) : null}
    <Dialog open={open} onOpenChange={close}>
      <DialogTrigger
        render={
          <Button variant="outline" type="button" disabled={disabled} />
        }
      >
        GitHub
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Publish to GitHub</DialogTitle>
          <DialogDescription>
            Creates a private repository from this version. The token is used
            once and is not saved.
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
        ) : null}
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
                .
              </p>
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
    </div>
  );
}
