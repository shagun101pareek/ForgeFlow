"use client";

import type { GeneratedFile } from "@/services/generation";

export function SourceView({
  files,
  activePath,
  onSelect,
  onChange,
}: {
  files: GeneratedFile[];
  activePath: string | null;
  onSelect: (path: string) => void;
  onChange: (path: string, code: string) => void;
}) {
  const active = files.find((file) => file.path === activePath) ?? files[0];

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex gap-1 overflow-x-auto border-b px-2 py-2">
        {files.map((file) => {
          const selected = file.path === active?.path;
          return (
            <button
              key={file.path}
              type="button"
              onClick={() => onSelect(file.path)}
              className={`rounded-md px-2.5 py-1 text-xs ${
                selected ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
              }`}
            >
              {file.path.replace(/^\//, "")}
            </button>
          );
        })}
      </div>
      <textarea
        aria-label={active ? active.path.replace(/^\//, "") : "Source"}
        value={active?.code ?? ""}
        onChange={(event) => {
          if (active) {
            onChange(active.path, event.target.value);
          }
        }}
        spellCheck={false}
        className="min-h-0 w-full flex-1 resize-none bg-transparent p-4 font-mono text-xs leading-5 outline-none"
      />
    </div>
  );
}
