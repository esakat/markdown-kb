import { renderMarkdown } from "../../lib/markdown";
import styles from "./MarkdownContent.module.css";
import "highlight.js/styles/github.css";

interface Props {
  source: string;
  docPath?: string;
}

export function MarkdownContent({ source, docPath }: Props) {
  const html = renderMarkdown(source, docPath);

  return (
    <div
      class={styles.prose}
      data-testid="markdown-content"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
