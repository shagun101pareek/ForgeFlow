import type { Metadata } from "next";

import { RequireAuth } from "@/features/auth/require-auth";
import { ProjectEditor } from "@/features/project/project-editor";

export const metadata: Metadata = {
  title: "Editor · ForgeFlow",
};

export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <RequireAuth>
      <ProjectEditor key={id} projectId={id} />
    </RequireAuth>
  );
}
