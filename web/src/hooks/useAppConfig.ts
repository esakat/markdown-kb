import { useState, useEffect } from "preact/hooks";
import { getConfig } from "../api/client";
import type { AppConfig } from "../types/api";

const DEFAULT_CONFIG: AppConfig = {
  title: "Markdown KB",
  theme: "default",
  themes: [],
  font: "default",
  font_url: "",
  font_family: "",
  tag_icons: [],
};

function applyAccent(theme: string) {
  if (typeof document === "undefined") return;
  const el = document.documentElement;
  if (theme && theme !== "default") {
    el.setAttribute("data-accent", theme);
  } else {
    el.removeAttribute("data-accent");
  }
}

function updateDocumentTitle(title: string) {
  if (typeof document !== "undefined") {
    document.title = title;
  }
}

function applyFont(fontUrl: string, fontFamily: string) {
  if (typeof document === "undefined") return;

  // Load Google Fonts stylesheet
  const existingLink = document.getElementById("kb-font-link");
  if (fontUrl) {
    if (existingLink) {
      (existingLink as HTMLLinkElement).href = fontUrl;
    } else {
      const link = document.createElement("link");
      link.id = "kb-font-link";
      link.rel = "stylesheet";
      link.href = fontUrl;
      document.head.appendChild(link);
    }
  } else if (existingLink) {
    existingLink.remove();
  }

  // Apply font-family via CSS variable
  if (fontFamily) {
    document.documentElement.style.setProperty("--font-sans", fontFamily);
  }
}

export function useAppConfig() {
  const [config, setConfig] = useState<AppConfig>(DEFAULT_CONFIG);

  useEffect(() => {
    getConfig()
      .then((cfg) => {
        setConfig(cfg);
        applyAccent(cfg.theme);
        updateDocumentTitle(cfg.title);
        applyFont(cfg.font_url || "", cfg.font_family || "");
      })
      .catch(() => {
        // Use defaults on error
      });
  }, []);

  return config;
}
