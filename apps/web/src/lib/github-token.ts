const TOKEN_STORAGE_KEY = "forgeflow.github.pat";

const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
    listener();
  }
}

export function subscribeGitHubToken(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function readGitHubToken() {
  if (typeof window === "undefined") {
    return "";
  }
  try {
    return window.localStorage.getItem(TOKEN_STORAGE_KEY) ?? "";
  } catch {
    return "";
  }
}

export function writeGitHubToken(value: string) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    const next = value.trim();
    if (next) {
      window.localStorage.setItem(TOKEN_STORAGE_KEY, next);
    } else {
      window.localStorage.removeItem(TOKEN_STORAGE_KEY);
    }
  } catch {
    // Ignore quota / private-mode failures.
  }
  emit();
}

export function clearGitHubToken() {
  writeGitHubToken("");
}

export function maskGitHubToken(token: string) {
  const value = token.trim();
  if (value.length < 8) {
    return value ? "••••••••" : "";
  }
  return `${value.slice(0, 4)}…${value.slice(-4)}`;
}
