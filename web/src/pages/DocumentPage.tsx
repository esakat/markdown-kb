import { useState, useEffect, useCallback } from "preact/hooks";
import { route } from "preact-router";
import { getDocument, getGraph } from "../api/client";
import type { DocumentDetail, GraphData } from "../types/api";
import { FrontmatterPanel } from "../components/Document/FrontmatterPanel";
import { MarkdownContent } from "../components/Document/MarkdownContent";
import { TableOfContents } from "../components/Document/TableOfContents";
import { CommitHistory } from "../components/Git/CommitHistory";
import { DiffView } from "../components/Git/DiffView";
import { MiniGraph } from "../components/Graph/MiniGraph";
import styles from "./DocumentPage.module.css";

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
  const [graph, setGraph] = useState<GraphData | null>(null);
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

    Promise.all([getDocument(docPath), getGraph()])
      .then(([docRes, graphRes]) => {
        if (cancelled) return;
        const raw = docRes as unknown as DocResponse;
        setDoc(raw.data);
        setGitDates(raw.git_dates);
        setGraph(graphRes.data);
        setLoading(false);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.message);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [docPath]);

  const handleNodeClick = useCallback((path: string) => {
    route(`/docs/${path}`);
  }, []);

  if (loading) return <p>Loading...</p>;
  if (error) return <p style={{ color: "#dc3545" }}>Error: {error}</p>;
  if (!doc) return <p>Document not found.</p>;

  const hasConnections =
    graph &&
    graph.edges.some(
      (e) => e.source === docPath || e.target === docPath,
    );

  return (
    <article>
      <h1>{doc.title || docPath}</h1>
      <FrontmatterPanel doc={doc} gitDates={gitDates} />

      <div class={styles.tocRow}>
        <div class={styles.tocLeft}>
          <TableOfContents markdown={doc.body} />
        </div>
        {hasConnections && graph && docPath && (
          <div class={styles.tocRight}>
            <MiniGraph
              centerPath={docPath}
              nodes={graph.nodes}
              edges={graph.edges}
              onNodeClick={handleNodeClick}
            />
          </div>
        )}
      </div>

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
