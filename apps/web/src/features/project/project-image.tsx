"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { getErrorMessage } from "@/lib/api-error";
import {
  deleteProjectImage,
  getProjectImage,
  uploadProjectImage,
} from "@/services/projects";

export function ProjectImage({
  projectId,
  hasImage,
}: {
  projectId: string;
  hasImage: boolean;
}) {
  const queryClient = useQueryClient();
  const [busy, setBusy] = useState(false);
  const image = useQuery({
    queryKey: ["project-image", projectId],
    queryFn: async () => URL.createObjectURL(await getProjectImage(projectId)),
    enabled: hasImage,
    retry: false,
  });

  useEffect(() => {
    const url = image.data;
    return () => {
      if (url) {
        URL.revokeObjectURL(url);
      }
    };
  }, [image.data]);

  async function onFile(file: File | undefined) {
    if (!file || busy) {
      return;
    }
    setBusy(true);
    try {
      const updated = await uploadProjectImage(projectId, file);
      queryClient.setQueryData(["project", projectId], updated);
      await queryClient.invalidateQueries({ queryKey: ["project-image", projectId] });
      toast.success("Image saved. Generate again to show it.");
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not save the image"));
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (busy) {
      return;
    }
    setBusy(true);
    try {
      const updated = await deleteProjectImage(projectId);
      queryClient.setQueryData(["project", projectId], updated);
      queryClient.removeQueries({ queryKey: ["project-image", projectId] });
      toast.success("Image removed. Generate again to update the preview.");
    } catch (error) {
      toast.error(getErrorMessage(error, "Could not remove the image"));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mb-6">
      <p className="mb-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
        Image
      </p>
      {hasImage && image.data ? (
        // Blob previews cannot go through the Next image optimizer.
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={image.data}
          alt=""
          className="mb-3 h-28 w-full rounded-lg object-cover"
        />
      ) : null}
      <div className="flex flex-wrap gap-2">
        <label className="inline-flex">
          <input
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            className="sr-only"
            disabled={busy}
            onChange={(event) => {
              const file = event.target.files?.[0];
              event.target.value = "";
              void onFile(file);
            }}
          />
          <span className="inline-flex h-8 cursor-pointer items-center rounded-lg border px-3 text-sm">
            {busy ? "Saving…" : hasImage ? "Replace" : "Add image"}
          </span>
        </label>
        {hasImage ? (
          <Button variant="outline" type="button" onClick={() => void remove()} disabled={busy}>
            Remove
          </Button>
        ) : null}
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        Shown in the hero the next time you generate.
      </p>
    </div>
  );
}
