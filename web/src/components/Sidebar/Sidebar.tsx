import { useTree } from "../../hooks/useTree";
import { TreeView } from "./TreeView";
import type { TagIcon } from "../../types/api";
import styles from "./Sidebar.module.css";

interface Props {
  currentPath?: string;
  tagIcons?: TagIcon[];
}

export function Sidebar({ currentPath, tagIcons }: Props) {
  const { tree, loading, error } = useTree();

  return (
    <aside class={styles.sidebar} data-testid="sidebar">
      <div class={styles.header}>
        <span class={styles.title}>Files</span>
      </div>
      <nav class={styles.nav}>
        {loading && <p class={styles.status}>Loading...</p>}
        {error && <p class={styles.error}>{error}</p>}
        {tree && (
          <TreeView tree={tree} currentPath={currentPath} tagIcons={tagIcons} />
        )}
      </nav>
    </aside>
  );
}
