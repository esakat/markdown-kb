import { useRef, useEffect } from "preact/hooks";
import * as d3 from "d3";
import type { GraphNode, GraphEdge } from "../../types/api";
import styles from "./MiniGraph.module.css";

interface Props {
  centerPath: string;
  nodes: GraphNode[];
  edges: GraphEdge[];
  onNodeClick?: (path: string) => void;
}

/**
 * MiniGraph renders a small ego-centric radial graph.
 * The current document is pinned at center, neighbors arranged in a circle.
 */
export function MiniGraph({ centerPath, nodes, edges, onNodeClick }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    // Extract 1-hop neighbors
    const neighborPaths = new Set<string>();
    const centerEdges: { target: string; type: string }[] = [];

    for (const e of edges) {
      if (e.source === centerPath && e.target !== centerPath) {
        if (!neighborPaths.has(e.target)) {
          neighborPaths.add(e.target);
          centerEdges.push({ target: e.target, type: e.type });
        }
      } else if (e.target === centerPath && e.source !== centerPath) {
        if (!neighborPaths.has(e.source)) {
          neighborPaths.add(e.source);
          centerEdges.push({ target: e.source, type: e.type });
        }
      }
    }

    if (centerEdges.length === 0) {
      svg
        .append("text")
        .attr("x", "50%")
        .attr("y", "50%")
        .attr("text-anchor", "middle")
        .attr("fill", "#999")
        .attr("font-size", "12px")
        .text("No connections");
      return;
    }

    const nodeMap = new Map<string, GraphNode>();
    for (const n of nodes) nodeMap.set(n.path, n);

    const width = svgRef.current.clientWidth || 340;
    const height = svgRef.current.clientHeight || 260;
    const cx = width / 2;
    const cy = height / 2;
    const radius = Math.min(cx, cy) - 50;

    const g = svg.append("g");

    // Position neighbors in a circle around center
    const neighborCount = centerEdges.length;

    interface PlacedNode {
      x: number;
      y: number;
      node: GraphNode;
      isCenter: boolean;
      angle: number;
    }

    const centerGraphNode = nodeMap.get(centerPath);
    const placed: PlacedNode[] = [];

    // Center node
    if (centerGraphNode) {
      placed.push({
        x: cx,
        y: cy,
        node: centerGraphNode,
        isCenter: true,
        angle: 0,
      });
    }

    // Neighbor nodes
    centerEdges.forEach((edge, i) => {
      const n = nodeMap.get(edge.target);
      if (!n) return;
      const angle = (2 * Math.PI * i) / neighborCount - Math.PI / 2;
      placed.push({
        x: cx + radius * Math.cos(angle),
        y: cy + radius * Math.sin(angle),
        node: n,
        isCenter: false,
        angle,
      });
    });

    // Draw edges (center -> each neighbor)
    for (const p of placed) {
      if (p.isCenter) continue;
      g.append("line")
        .attr("x1", cx)
        .attr("y1", cy)
        .attr("x2", p.x)
        .attr("y2", p.y)
        .attr("stroke", "#ccc")
        .attr("stroke-width", 1)
        .attr("stroke-opacity", 0.6);
    }

    // Draw nodes
    for (const p of placed) {
      const group = g
        .append("g")
        .attr("transform", `translate(${p.x},${p.y})`)
        .style("cursor", p.isCenter ? "default" : "pointer");

      if (!p.isCenter) {
        group.on("click", () => onNodeClick?.(p.node.path));
      }

      const r = p.isCenter ? 8 : 5;
      const fill = p.isCenter
        ? "var(--color-primary, #0066cc)"
        : p.node.tags.length === 0
          ? "#999"
          : d3.schemeTableau10[
              Math.abs(hashCode(p.node.tags[0])) %
                d3.schemeTableau10.length
            ];

      group
        .append("circle")
        .attr("r", r)
        .attr("fill", fill)
        .attr("stroke", "#fff")
        .attr("stroke-width", p.isCenter ? 2.5 : 1.5);

      // Label: position based on angle to avoid overlap with edges
      const label = truncate(
        p.node.title || p.node.path.replace(/\.md$/, ""),
        16,
      );

      if (p.isCenter) {
        // Center label below the node
        group
          .append("text")
          .text(label)
          .attr("text-anchor", "middle")
          .attr("dy", r + 14)
          .attr("font-size", "10px")
          .attr("font-weight", "bold")
          .attr("fill", "var(--color-primary, #0066cc)")
          .attr("pointer-events", "none");
      } else {
        // Neighbor labels: anchor based on which side of center
        const onLeft = p.x < cx - 10;
        const onRight = p.x > cx + 10;
        const anchor = onLeft ? "end" : onRight ? "start" : "middle";
        const dx = onLeft ? -9 : onRight ? 9 : 0;
        const dy = !onLeft && !onRight ? (p.y < cy ? -10 : 16) : 4;

        group
          .append("text")
          .text(label)
          .attr("text-anchor", anchor)
          .attr("dx", dx)
          .attr("dy", dy)
          .attr("font-size", "9px")
          .attr("fill", "var(--color-text-secondary, #666)")
          .attr("pointer-events", "none");
      }

      // Tooltip with full title
      group.append("title").text(p.node.title || p.node.path);
    }

    return () => {};
  }, [centerPath, nodes, edges, onNodeClick]);

  return <svg ref={svgRef} class={styles.svg} data-testid="mini-graph" />;
}

function hashCode(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) - hash + str.charCodeAt(i);
    hash |= 0;
  }
  return hash;
}

function truncate(str: string, max: number): string {
  return str.length > max ? str.slice(0, max) + "..." : str;
}
