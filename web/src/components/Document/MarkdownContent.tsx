import { useRef, useEffect } from "preact/hooks";
import mermaid from "mermaid";
import { renderMarkdown } from "../../lib/markdown";
import styles from "./MarkdownContent.module.css";
import "highlight.js/styles/github.css";

interface Props {
  source: string;
  docPath?: string;
}

function getCurrentTheme(): "default" | "dark" {
  return document.documentElement.getAttribute("data-theme") === "dark"
    ? "dark"
    : "default";
}

export function MarkdownContent({ source, docPath }: Props) {
  const html = renderMarkdown(source, docPath);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ref.current) return;
    const mermaidEls = ref.current.querySelectorAll<HTMLElement>("pre.mermaid");
    if (mermaidEls.length > 0) {
      mermaid.initialize({
        startOnLoad: false,
        theme: getCurrentTheme(),
        securityLevel: "loose",
      });
      // Reset processed state so mermaid re-renders on navigation
      mermaidEls.forEach((el) => el.removeAttribute("data-processed"));
      mermaid.run({ nodes: mermaidEls });
    }
  }, [html]);

  return (
    <div
      ref={ref}
      class={styles.prose}
      data-testid="markdown-content"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
