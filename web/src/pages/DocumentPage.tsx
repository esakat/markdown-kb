import { useState, useEffect } from "preact/hooks";
import { getDocument } from "../api/client";
import type { DocumentDetail } from "../types/api";
import { FrontmatterPanel } from "../components/Document/FrontmatterPanel";
import { MarkdownContent } from "../components/Document/MarkdownContent";
import { TableOfContents } from "../components/Document/TableOfContents";

interface Props {
  path?: string;
  docPath?: string;
}

export function DocumentPage({ docPath }: Props) {
  const [doc, setDoc] = useState<DocumentDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!docPath) return;
    let cancelled = false;
    setLoading(true);
    setError(null);

    getDocument(docPath)
      .then((res) => {
        if (!cancelled) {
          setDoc(res.data);
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
      <FrontmatterPanel doc={doc} />
      <TableOfContents markdown={doc.body} />
      <MarkdownContent source={doc.body} docPath={docPath} />
    </article>
  );
}
