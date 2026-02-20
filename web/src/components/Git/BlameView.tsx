import { useState, useEffect, useRef } from "preact/hooks";
import { getFileBlame } from "../../api/client";
import type { BlameLine } from "../../types/api";
import styles from "./BlameView.module.css";

interface Props {
  docPath: string;
  lineNo: number;
  anchorRect: { top: number; left: number };
  onClose: () => void;
}

export function BlamePopover({ docPath, lineNo, anchorRect, onClose }: Props) {
  const [blame, setBlame] = useState<BlameLine | null>(null);
  const [loading, setLoading] = useState(true);
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    getFileBlame(docPath, lineNo, lineNo)
      .then((res) => {
        if (!cancelled && res.data.length > 0) {
          setBlame(res.data[0]);
          setLoading(false);
        } else if (!cancelled) {
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [docPath, lineNo]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("ja-JP", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    });
  };

  return (
    <>
      <div class={styles.overlay} onClick={onClose} />
      <div
        ref={popoverRef}
        class={styles.popover}
        style={{
          top: `${anchorRect.top}px`,
          left: `${anchorRect.left}px`,
        }}
        data-testid="blame-popover"
      >
        <div class={styles.header}>
          <span class={styles.title}>Blame - Line {lineNo}</span>
          <button class={styles.closeBtn} onClick={onClose} aria-label="Close">
            &#x2715;
          </button>
        </div>

        {loading && <div class={styles.loading}>Loading blame...</div>}

        {!loading && blame && (
          <>
            <div class={styles.commitHash}>{blame.hash.substring(0, 7)}</div>
            <div class={styles.commitMeta}>
              <span>{blame.author}</span>
              <span>{formatDate(blame.date)}</span>
            </div>
          </>
        )}

        {!loading && !blame && (
          <div class={styles.loading}>No blame data available.</div>
        )}
      </div>
    </>
  );
}
