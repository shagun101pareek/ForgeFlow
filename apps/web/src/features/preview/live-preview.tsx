"use client";

import {
  SandpackLayout,
  SandpackPreview,
  SandpackProvider,
} from "@codesandbox/sandpack-react";
import { useTheme } from "next-themes";

import type { GeneratedFile } from "@/services/generation";

export function LivePreview({
  files,
  route,
}: {
  files: GeneratedFile[];
  route: string;
}) {
  const { resolvedTheme } = useTheme();
  const prepared = files.map((file) =>
    file.path === "/App.js"
      ? {
          ...file,
          code: file.code.replace(
            /const INITIAL_ROUTE = ".*?";/,
            `const INITIAL_ROUTE = ${JSON.stringify(route)};`,
          ),
        }
      : file,
  );
  const sandpackFiles = Object.fromEntries(
    prepared.map((file) => [file.path, file.code]),
  );

  return (
    <SandpackProvider
      template="react"
      theme={resolvedTheme === "dark" ? "dark" : "light"}
      files={sandpackFiles}
    >
      <SandpackLayout style={{ height: "100%", minHeight: 560, border: 0 }}>
        <SandpackPreview
          style={{ height: 560 }}
          showOpenInCodeSandbox={false}
          showRefreshButton
        />
      </SandpackLayout>
    </SandpackProvider>
  );
}
