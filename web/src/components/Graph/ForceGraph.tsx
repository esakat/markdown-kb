import { useRef, useEffect } from "preact/hooks";
import * as d3 from "d3";
import type { GraphNode, GraphEdge } from "../../types/api";
import styles from "./ForceGraph.module.css";

interface Props {
  nodes: GraphNode[];
  edges: GraphEdge[];
  onNodeClick?: (path: string) => void;
}

interface SimNode extends d3.SimulationNodeDatum {
  path: string;
  title: string;
  tags: string[];
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  type: string;
  label?: string;
}

export function ForceGraph({ nodes, edges, onNodeClick }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current || nodes.length === 0) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const width = svgRef.current.clientWidth || 800;
    const height = svgRef.current.clientHeight || 600;

    // Build simulation data
    const simNodes: SimNode[] = nodes.map((n) => ({
      ...n,
      x: undefined,
      y: undefined,
    }));

    const nodeMap = new Map<string, SimNode>();
    simNodes.forEach((n) => nodeMap.set(n.path, n));

    const simLinks: SimLink[] = edges
      .filter((e) => nodeMap.has(e.source) && nodeMap.has(e.target))
      .map((e) => ({
        source: nodeMap.get(e.source)!,
        target: nodeMap.get(e.target)!,
        type: e.type,
        label: e.label,
      }));

    // Create zoom container
    const g = svg.append("g");

    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on("zoom", (event) => {
        g.attr("transform", event.transform);
      });

    svg.call(zoom);

    // Create simulation
    const simulation = d3
      .forceSimulation(simNodes)
      .force(
        "link",
        d3
          .forceLink<SimNode, SimLink>(simLinks)
          .id((d) => d.path)
          .distance(100),
      )
      .force("charge", d3.forceManyBody().strength(-200))
      .force("center", d3.forceCenter(width / 2, height / 2))
      .force("collision", d3.forceCollide().radius(30));

    // Draw edges
    const link = g
      .append("g")
      .selectAll("line")
      .data(simLinks)
      .join("line")
      .attr("class", (d) =>
        d.type === "link" ? styles.linkEdge : styles.tagEdge,
      )
      .attr("stroke-width", 1.5);

    // Draw nodes
    const node = g
      .append("g")
      .selectAll<SVGGElement, SimNode>("g")
      .data(simNodes)
      .join("g")
      .attr("class", styles.node)
      .style("cursor", "pointer")
      .on("click", (_event, d) => {
        onNodeClick?.(d.path);
      })
      .call(
        d3
          .drag<SVGGElement, SimNode>()
          .on("start", (event, d) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on("drag", (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on("end", (event, d) => {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }),
      );

    node
      .append("circle")
      .attr("r", (d) => 6 + d.tags.length * 2)
      .attr("fill", (d) => {
        if (d.tags.length === 0) return "#999";
        return d3.schemeTableau10[
          Math.abs(hashCode(d.tags[0])) % d3.schemeTableau10.length
        ];
      });

    node
      .append("text")
      .text((d) => d.title || d.path.replace(/\.md$/, ""))
      .attr("dx", 12)
      .attr("dy", 4)
      .attr("class", styles.label);

    // Tooltip
    node.append("title").text((d) => {
      const tags = d.tags.length > 0 ? `\nTags: ${d.tags.join(", ")}` : "";
      return `${d.title || d.path}${tags}`;
    });

    simulation.on("tick", () => {
      link
        .attr("x1", (d) => (d.source as SimNode).x!)
        .attr("y1", (d) => (d.source as SimNode).y!)
        .attr("x2", (d) => (d.target as SimNode).x!)
        .attr("y2", (d) => (d.target as SimNode).y!);

      node.attr("transform", (d) => `translate(${d.x},${d.y})`);
    });

    return () => {
      simulation.stop();
    };
  }, [nodes, edges, onNodeClick]);

  return (
    <svg ref={svgRef} class={styles.svg} data-testid="force-graph">
      {nodes.length === 0 && (
        <text x="50%" y="50%" text-anchor="middle" fill="#999">
          No documents to display
        </text>
      )}
    </svg>
  );
}

function hashCode(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) - hash + str.charCodeAt(i);
    hash |= 0;
  }
  return hash;
}
