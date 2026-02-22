import type { TreeNode, TagIcon } from "../../types/api";
import { TreeNodeItem } from "./TreeNode";
import styles from "./TreeView.module.css";

interface Props {
  tree: TreeNode;
  currentPath?: string;
  tagIcons?: TagIcon[];
}

export function TreeView({ tree, currentPath, tagIcons }: Props) {
  if (!tree.children || tree.children.length === 0) {
    return <p class={styles.empty}>No documents found.</p>;
  }

  return (
    <ul class={styles.root} data-testid="tree-view">
      {tree.children.map((node) => (
        <TreeNodeItem
          key={node.name}
          node={node}
          currentPath={currentPath}
          tagIcons={tagIcons}
        />
      ))}
    </ul>
  );
}
