import { useState } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { Sidebar } from "./Sidebar/Sidebar";
import styles from "./Layout.module.css";

interface Props {
  currentPath?: string;
  children: ComponentChildren;
}

export function Layout({ currentPath, children }: Props) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div class={styles.layout}>
      <header class={styles.header}>
        <button
          class={styles.hamburger}
          onClick={() => setSidebarOpen(!sidebarOpen)}
          aria-label="Toggle sidebar"
          data-testid="hamburger"
        >
          ☰
        </button>
        <h1 class={styles.title}>Markdown KB</h1>
      </header>
      <div class={styles.body}>
        <div class={`${styles.sidebarWrapper} ${sidebarOpen ? styles.open : ""}`}>
          <Sidebar currentPath={currentPath} />
        </div>
        {sidebarOpen && (
          <div
            class={styles.overlay}
            onClick={() => setSidebarOpen(false)}
          />
        )}
        <main class={styles.main}>{children}</main>
      </div>
    </div>
  );
}
