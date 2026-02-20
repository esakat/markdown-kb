import { useState, useEffect } from "preact/hooks";
import { getFileHistory } from "../../api/client";
import type { GitCommit } from "../../types/api";
import styles from "./CommitHistory.module.css";

interface Props {
  docPath: string;
  onSelectDiff?: (from: string, to: string) => void;
}

export function CommitHistory({ docPath, onSelectDiff }: Props) {
  const [commits, setCommits] = useState<GitCommit[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    getFileHistory(docPath)
      .then((res) => {
        if (!cancelled) {
          setCommits(res.data || []);
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setCommits([]);
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [docPath]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("ja-JP", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  };

  const handleDiffClick = (index: number) => {
    if (index < commits.length - 1 && onSelectDiff) {
      onSelectDiff(commits[index + 1].hash, commits[index].hash);
    }
  };

  return (
    <div class={styles.panel} data-testid="commit-history">
      <div class={styles.header} onClick={() => setOpen(!open)}>
        <span class={styles.headerTitle}>
          Commit History
          {!loading && ` (${commits.length})`}
        </span>
        <span class={`${styles.toggle} ${open ? styles.toggleOpen : ""}`}>
          &#9654;
        </span>
      </div>

      {open && (
        <div class={styles.list}>
          {loading && <div class={styles.loading}>Loading history...</div>}
          {!loading && commits.length === 0 && (
            <div class={styles.empty}>No commit history available.</div>
          )}
          {!loading &&
            commits.map((commit, i) => (
              <div key={commit.hash} class={styles.commitItem}>
                <div class={styles.commitDot} />
                <div class={styles.commitBody}>
                  <div class={styles.commitMessage}>{commit.message}</div>
                  <div class={styles.commitMeta}>
                    <span>{commit.author}</span>
                    <span>{formatDate(commit.date)}</span>
                    <span class={styles.commitHash}>
                      {commit.hash.substring(0, 7)}
                    </span>
                  </div>
                </div>
                {i < commits.length - 1 && onSelectDiff && (
                  <button
                    class={styles.diffBtn}
                    onClick={() => handleDiffClick(i)}
                    title="Show diff with previous commit"
                  >
                    diff
                  </button>
                )}
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
