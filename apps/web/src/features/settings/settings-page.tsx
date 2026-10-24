"use client";

import Link from "next/link";
import { useState, useSyncExternalStore } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Wordmark } from "@/components/wordmark";
import { AccountMenu } from "@/features/dashboard/account-menu";
import {
  clearGitHubToken,
  maskGitHubToken,
  readGitHubToken,
  subscribeGitHubToken,
  writeGitHubToken,
} from "@/lib/github-token";
import { useAuthStore } from "@/store/auth";

export function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const savedToken = useSyncExternalStore(
    subscribeGitHubToken,
    readGitHubToken,
    () => "",
  );
  const [draftToken, setDraftToken] = useState("");

  function saveToken() {
    const next = draftToken.trim();
    if (next.length < 20) {
      toast.error("Paste a GitHub personal access token with the repo scope");
      return;
    }
    writeGitHubToken(next);
    setDraftToken("");
    toast.success("Token saved on this device");
  }

  function removeToken() {
    clearGitHubToken();
    setDraftToken("");
    toast.success("Saved token cleared");
  }

  return (
    <div className="flex flex-1 flex-col">
      <header className="flex items-center justify-between gap-4 border-b px-4 py-3 sm:px-6">
        <Wordmark href="/dashboard" />
        <AccountMenu />
      </header>
      <main className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-10 px-4 py-8 sm:px-6">
        <div>
          <p className="text-sm text-muted-foreground">
            <Link href="/dashboard" className="underline-offset-4 hover:underline">
              Projects
            </Link>
            <span className="mx-2 text-foreground/30">/</span>
            Settings
          </p>
          <h1 className="mt-2 font-heading text-4xl font-medium tracking-tight">
            Settings
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Account details and the GitHub token used when you publish a
            preview.
          </p>
        </div>

        <section className="space-y-4">
          <h2 className="font-heading text-2xl font-medium">Account</h2>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                Name
              </dt>
              <dd className="mt-1 text-sm">{user?.name ?? "—"}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                Email
              </dt>
              <dd className="mt-1 text-sm">{user?.email ?? "—"}</dd>
            </div>
          </dl>
        </section>

        <section className="space-y-4">
          <div>
            <h2 className="font-heading text-2xl font-medium">GitHub token</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Stored only in this browser. Used by Publish in the editor. Needs
              the <span className="font-medium text-foreground">repo</span>{" "}
              scope.
            </p>
          </div>
          <p className="text-sm">
            {savedToken ? (
              <>
                Saved token:{" "}
                <span className="font-mono text-muted-foreground">
                  {maskGitHubToken(savedToken)}
                </span>
              </>
            ) : (
              <span className="text-muted-foreground">No token saved yet.</span>
            )}
          </p>
          <div className="flex flex-col gap-2">
            <Label htmlFor="settings-github-token">
              {savedToken ? "Replace token" : "Personal access token"}
            </Label>
            <Input
              id="settings-github-token"
              type="password"
              autoComplete="off"
              placeholder="ghp_…"
              value={draftToken}
              onChange={(event) => setDraftToken(event.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Create one at{" "}
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
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              onClick={saveToken}
              disabled={draftToken.trim().length < 20}
            >
              Save on this device
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={removeToken}
              disabled={!savedToken}
            >
              Clear saved token
            </Button>
          </div>
        </section>
      </main>
    </div>
  );
}
