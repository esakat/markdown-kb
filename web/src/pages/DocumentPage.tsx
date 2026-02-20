import { useState, useEffect } from "preact/hooks";
import { getDocument } from "../api/client";
import type { DocumentDetail } from "../types/api";
import { FrontmatterPanel } from "../components/Document/FrontmatterPanel";
import { MarkdownContent } from "../components/Document/MarkdownContent";
import { TableOfContents } from "../components/Document/TableOfContents";
import { CommitHistory } from "../components/Git/CommitHistory";
import { DiffView } from "../components/Git/DiffView";

interface Props {
  path?: string;
  docPath?: string;
}

interface DocResponse {
  data: DocumentDetail;
  git_dates?: { created?: string; updated?: string };
}

export function DocumentPage({ docPath }: Props) {
  const [doc, setDoc] = useState<DocumentDetail | null>(null);
  const [gitDates, setGitDates] = useState<
    { created?: string; updated?: string } | undefined
  >(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [diffRange, setDiffRange] = useState<{
    from: string;
    to: string;
  } | null>(null);

  useEffect(() => {
    if (!docPath) return;
    let cancelled = false;
    setLoading(true);
    setError(null);
    setDiffRange(null);

    // Fetch the raw response to capture git_dates alongside data
    getDocument(docPath)
      .then((res) => {
        if (!cancelled) {
          const raw = res as unknown as DocResponse;
          setDoc(raw.data);
          setGitDates(raw.git_dates);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.message);
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [docPath]);

  if (loading) return <p>Loading...</p>;
  if (error) return <p style={{ color: "#dc3545" }}>Error: {error}</p>;
  if (!doc) return <p>Document not found.</p>;

  return (
    <article>
      <h1>{doc.title || docPath}</h1>
      <FrontmatterPanel doc={doc} gitDates={gitDates} />
      <TableOfContents markdown={doc.body} />
      <MarkdownContent source={doc.body} docPath={docPath} />

      {docPath && (
        <CommitHistory
          docPath={docPath}
          onSelectDiff={(from, to) => setDiffRange({ from, to })}
        />
      )}

      {diffRange && docPath && (
        <DiffView
          docPath={docPath}
          fromHash={diffRange.from}
          toHash={diffRange.to}
          onClose={() => setDiffRange(null)}
        />
      )}
    </article>
  );
}
