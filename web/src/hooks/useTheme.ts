import { useState, useEffect, useCallback } from "preact/hooks";

type Theme = "light" | "dark";

const STORAGE_KEY = "markdown-kb:theme";

function getStorage(): Storage | null {
  try {
    if (
      typeof localStorage !== "undefined" &&
      typeof localStorage.getItem === "function"
    ) {
      return localStorage;
    }
    return null;
  } catch {
    return null;
  }
}

function getInitialTheme(): Theme {
  const storage = getStorage();
  const stored = storage?.getItem(STORAGE_KEY) as Theme | null;
  if (stored === "light" || stored === "dark") return stored;
  if (typeof window !== "undefined" && window.matchMedia) {
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }
  return "light";
}

function applyTheme(theme: Theme) {
  if (typeof document !== "undefined") {
    document.documentElement.setAttribute("data-theme", theme);
  }
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme);

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  // Listen for OS preference changes when no explicit preference is stored
  useEffect(() => {
    if (typeof window === "undefined" || !window.matchMedia) return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = (e: MediaQueryListEvent) => {
      const storage = getStorage();
      if (!storage?.getItem(STORAGE_KEY)) {
        const next = e.matches ? "dark" : "light";
        setThemeState(next);
      }
    };
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => {
      const next = prev === "light" ? "dark" : "light";
      getStorage()?.setItem(STORAGE_KEY, next);
      return next;
    });
  }, []);

  return { theme, toggleTheme } as const;
}
