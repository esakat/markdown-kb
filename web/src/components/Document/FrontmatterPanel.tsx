import type { DocumentDetail } from "../../types/api";
import { TagBadge } from "./TagBadge";
import { StatusBadge } from "./StatusBadge";
import styles from "./FrontmatterPanel.module.css";

interface Props {
  doc: DocumentDetail;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  return d.toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  return `${(bytes / 1024).toFixed(1)} KB`;
}

export function FrontmatterPanel({ doc }: Props) {
  const status = doc.meta?.status as string | undefined;
  const tags = (doc.meta?.tags as string[]) ?? [];

  return (
    <div class={styles.panel} data-testid="frontmatter-panel">
      <div class={styles.badges}>
        {status && <StatusBadge status={status} />}
        {tags.map((tag) => (
          <TagBadge key={tag} tag={tag} />
        ))}
      </div>
      <div class={styles.meta}>
        <span class={styles.metaItem}>
          {formatDate(doc.mod_time)}
        </span>
        <span class={styles.metaItem}>{formatSize(doc.size)}</span>
      </div>
    </div>
  );
}
