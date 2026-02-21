import { useState, useEffect, useCallback } from "preact/hooks";
import { route } from "preact-router";
import { getGraph, listTags } from "../api/client";
import type { GraphData, TagCount } from "../types/api";
import { ForceGraph } from "../components/Graph/ForceGraph";
import { TagCloud } from "../components/Graph/TagCloud";
import styles from "./GraphPage.module.css";

interface Props {
  path?: string;
}

export function GraphPage({}: Props) {
  const [graph, setGraph] = useState<GraphData | null>(null);
  const [tags, setTags] = useState<TagCount[]>([]);
  const [selectedTag, setSelectedTag] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    Promise.all([getGraph(), listTags()])
      .then(([graphRes, tagsRes]) => {
        if (cancelled) return;
        setGraph(graphRes.data);
        setTags(tagsRes.data);
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
  }, []);

  const handleNodeClick = useCallback((path: string) => {
    route(`/docs/${path}`);
  }, []);

  if (loading) return <p>Loading graph...</p>;
  if (error) return <p style={{ color: "#dc3545" }}>Error: {error}</p>;
  if (!graph) return <p>No graph data.</p>;

  // Filter graph by selected tag
  const filteredNodes = selectedTag
    ? graph.nodes.filter((n) => n.tags && n.tags.includes(selectedTag))
    : graph.nodes;

  const filteredPaths = new Set(filteredNodes.map((n) => n.path));
  const filteredEdges = graph.edges.filter(
    (e) => filteredPaths.has(e.source) && filteredPaths.has(e.target),
  );

  return (
    <div class={styles.page}>
      <h1>Document Graph</h1>
      <p class={styles.subtitle}>
        {graph.nodes.length} documents, {graph.edges.length} connections
      </p>

      <div class={styles.layout}>
        <div class={styles.sidebar}>
          <h2>Tags</h2>
          <TagCloud
            tags={tags}
            selectedTag={selectedTag}
            onTagClick={setSelectedTag}
          />
        </div>
        <div class={styles.graphArea}>
          <ForceGraph
            nodes={filteredNodes}
            edges={filteredEdges}
            onNodeClick={handleNodeClick}
          />
          <div class={styles.legend}>
            <span class={styles.legendItem}>
              <span
                class={styles.legendLine}
                style={{ borderStyle: "dashed", borderColor: "#0066cc" }}
              />
              Link
            </span>
            <span class={styles.legendItem}>
              <span
                class={styles.legendLine}
                style={{ borderStyle: "solid", borderColor: "#999" }}
              />
              Shared tag
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
