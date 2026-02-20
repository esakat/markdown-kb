import { useState, useEffect } from "preact/hooks";
import { getFileDiff } from "../../api/client";
import styles from "./DiffView.module.css";

interface Props {
  docPath: string;
  fromHash: string;
  toHash: string;
  onClose: () => void;
}

interface DiffLine {
  type: "add" | "del" | "context" | "hunk" | "header";
  content: string;
  lineNo?: number;
}

export function DiffView({ docPath, fromHash, toHash, onClose }: Props) {
  const [lines, setLines] = useState<DiffLine[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    getFileDiff(docPath, fromHash, toHash)
      .then((res) => {
        if (!cancelled) {
          setLines(parseDiff(res.data));
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setLines([]);
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [docPath, fromHash, toHash]);

  return (
    <div class={styles.container} data-testid="diff-view">
      <div class={styles.header}>
        <div>
          <span class={styles.headerTitle}>Diff </span>
          <span class={styles.headerHashes}>
            {fromHash.substring(0, 7)}..{toHash.substring(0, 7)}
          </span>
        </div>
        <button class={styles.closeBtn} onClick={onClose} aria-label="Close diff">
          &#x2715;
        </button>
      </div>
      <div class={styles.diffContent}>
        {loading && <div class={styles.loading}>Loading diff...</div>}
        {!loading && lines.length === 0 && (
          <div class={styles.empty}>No changes between these commits.</div>
        )}
        {!loading &&
          lines.map((line, i) => {
            const lineStyle =
              line.type === "add"
                ? styles.lineAdd
                : line.type === "del"
                  ? styles.lineDel
                  : line.type === "hunk"
                    ? styles.lineHunk
                    : "";

            return (
              <div key={i} class={`${styles.line} ${lineStyle}`}>
                <span class={styles.lineNo}>
                  {line.lineNo !== undefined ? line.lineNo : ""}
                </span>
                <span class={styles.lineContent}>{line.content}</span>
              </div>
            );
          })}
      </div>
    </div>
  );
}

function parseDiff(raw: string): DiffLine[] {
  if (!raw) return [];

  const lines: DiffLine[] = [];
  let lineNo = 0;

  for (const line of raw.split("\n")) {
    if (line.startsWith("@@")) {
      // Parse hunk header: @@ -a,b +c,d @@
      const match = line.match(/\+(\d+)/);
      if (match) {
        lineNo = parseInt(match[1], 10);
      }
      lines.push({ type: "hunk", content: line });
    } else if (line.startsWith("+++") || line.startsWith("---")) {
      lines.push({ type: "header", content: line });
    } else if (line.startsWith("+")) {
      lines.push({ type: "add", content: line.substring(1), lineNo });
      lineNo++;
    } else if (line.startsWith("-")) {
      lines.push({ type: "del", content: line.substring(1) });
    } else if (line.startsWith("diff") || line.startsWith("index")) {
      lines.push({ type: "header", content: line });
    } else {
      lines.push({ type: "context", content: line.substring(1) || "", lineNo });
      lineNo++;
    }
  }

  return lines;
}
