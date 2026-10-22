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
  height = 560,
  showRefreshButton = true,
  showNavigator = false,
  theme: themeProp,
  className,
}: {
  files: GeneratedFile[];
  route: string;
  height?: number | string;
  showRefreshButton?: boolean;
  showNavigator?: boolean;
  theme?: "light" | "dark";
  className?: string;
}) {
  const { resolvedTheme } = useTheme();
  const theme =
    themeProp ?? (resolvedTheme === "dark" ? "dark" : "light");
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
  const frameHeight =
    typeof height === "number" ? height : undefined;

  return (
    <div
      className={className}
      style={{
        height: typeof height === "string" ? height : frameHeight,
        minHeight: frameHeight,
      }}
    >
      <SandpackProvider template="react" theme={theme} files={sandpackFiles}>
        <SandpackLayout
          style={{
            height: "100%",
            minHeight: frameHeight,
            border: 0,
            borderRadius: 0,
          }}
        >
          <SandpackPreview
            style={{ height: "100%", minHeight: frameHeight }}
            showOpenInCodeSandbox={false}
            showRefreshButton={showRefreshButton}
            showNavigator={showNavigator}
          />
        </SandpackLayout>
      </SandpackProvider>
    </div>
  );
}
