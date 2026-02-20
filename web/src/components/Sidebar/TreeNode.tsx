import { useState } from "preact/hooks";
import { route } from "preact-router";
import type { TreeNode as TreeNodeType } from "../../types/api";
import styles from "./TreeNode.module.css";

interface Props {
  node: TreeNodeType;
  currentPath?: string;
  depth?: number;
}

export function TreeNodeItem({ node, currentPath, depth = 0 }: Props) {
  const [expanded, setExpanded] = useState(depth < 1);
  const isDir = node.type === "dir";
  const isActive = !isDir && node.path === currentPath;

  const handleClick = () => {
    if (isDir) {
      setExpanded(!expanded);
    } else if (node.path) {
      route(`/docs/${node.path}`);
    }
  };

  return (
    <li class={styles.item}>
      <button
        class={`${styles.label} ${isActive ? styles.active : ""}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={handleClick}
        data-testid={`tree-node-${node.name}`}
        title={node.title || node.name}
      >
        <span class={styles.icon}>{isDir ? (expanded ? "▾" : "▸") : "📄"}</span>
        <span class={styles.name}>{node.title && !isDir ? node.title : node.name}</span>
      </button>
      {isDir && expanded && node.children && (
        <ul class={styles.children}>
          {node.children.map((child) => (
            <TreeNodeItem
              key={child.name}
              node={child}
              currentPath={currentPath}
              depth={depth + 1}
            />
          ))}
        </ul>
      )}
    </li>
  );
}
