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

export function PublishDialog({
  projectId,
  generationId,
  disabled,
}: {
  projectId: string;
  generationId: string | null;
  disabled: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [url, setUrl] = useState<string | null>(null);

  function close(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setToken("");
      setUrl(null);
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
      setToken("");
      setUrl(result.url);
      toast.success("Repository created");
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not publish to GitHub"));
    } finally {
      setPublishing(false);
    }
  }

  return (
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
        ) : (
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
  );
}
