import type { TreeNode } from "../../types/api";
import { TreeNodeItem } from "./TreeNode";
import styles from "./TreeView.module.css";

interface Props {
  tree: TreeNode;
  currentPath?: string;
}

export function TreeView({ tree, currentPath }: Props) {
  if (!tree.children || tree.children.length === 0) {
    return <p class={styles.empty}>No documents found.</p>;
  }

  return (
    <ul class={styles.root} data-testid="tree-view">
      {tree.children.map((node) => (
        <TreeNodeItem key={node.name} node={node} currentPath={currentPath} />
      ))}
    </ul>
  );
}
