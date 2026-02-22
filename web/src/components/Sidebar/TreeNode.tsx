import { useState } from "preact/hooks";
import { route } from "preact-router";
import type { TreeNode as TreeNodeType, TagIcon } from "../../types/api";
import styles from "./TreeNode.module.css";

interface Props {
  node: TreeNodeType;
  currentPath?: string;
  depth?: number;
  tagIcons?: TagIcon[];
}

function getFileIcon(
  tags: string[] | undefined,
  tagIcons: TagIcon[] | undefined,
): string | null {
  if (!tags || !tagIcons || tagIcons.length === 0) return null;
  for (const ti of tagIcons) {
    if (tags.includes(ti.tag)) return ti.emoji;
  }
  return null;
}

export function TreeNodeItem({
  node,
  currentPath,
  depth = 0,
  tagIcons,
}: Props) {
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

  const fileEmoji = !isDir ? getFileIcon(node.tags, tagIcons) : null;

  return (
    <li class={styles.item}>
      <button
        class={`${styles.label} ${isActive ? styles.active : ""}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={handleClick}
        data-testid={`tree-node-${node.name}`}
        title={node.title || node.name}
      >
        {isDir ? (
          <span
            class={`${styles.chevron} ${expanded ? styles.chevronOpen : ""}`}
          />
        ) : fileEmoji ? (
          <span class={styles.fileEmoji}>{fileEmoji}</span>
        ) : null}
        <span class={styles.name}>
          {node.title && !isDir ? node.title : node.name}
        </span>
      </button>
      {isDir && expanded && node.children && (
        <ul class={styles.children}>
          {node.children.map((child) => (
            <TreeNodeItem
              key={child.name}
              node={child}
              currentPath={currentPath}
              depth={depth + 1}
              tagIcons={tagIcons}
            />
          ))}
        </ul>
      )}
    </li>
  );
}
