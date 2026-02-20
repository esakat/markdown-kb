import type { TagCount } from "../../types/api";
import styles from "./TagCloud.module.css";

interface Props {
  tags: TagCount[];
  selectedTag?: string;
  onTagClick?: (tag: string | undefined) => void;
}

export function TagCloud({ tags, selectedTag, onTagClick }: Props) {
  if (tags.length === 0) {
    return <p class={styles.empty}>No tags found.</p>;
  }

  const maxCount = Math.max(...tags.map((t) => t.count));

  return (
    <div class={styles.cloud} data-testid="tag-cloud">
      {tags.map((t) => {
        const size = 0.75 + (t.count / maxCount) * 0.75;
        const isActive = selectedTag === t.tag;
        return (
          <button
            key={t.tag}
            class={`${styles.tag} ${isActive ? styles.active : ""}`}
            style={{ fontSize: `${size}rem` }}
            onClick={() => onTagClick?.(isActive ? undefined : t.tag)}
            title={`${t.tag} (${t.count})`}
          >
            {t.tag}
            <span class={styles.count}>{t.count}</span>
          </button>
        );
      })}
    </div>
  );
}
