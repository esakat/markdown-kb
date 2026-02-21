import { useState } from "preact/hooks";
import type { ComponentChildren } from "preact";
import { Sidebar } from "./Sidebar/Sidebar";
import { SearchBar } from "./Search/SearchBar";
import { ThemeToggle } from "./ThemeToggle";
import { useTheme } from "../hooks/useTheme";
import { useAppConfig } from "../hooks/useAppConfig";
import styles from "./Layout.module.css";

interface Props {
  currentPath?: string;
  onSearch?: (query: string) => void;
  children: ComponentChildren;
}

export function Layout({ currentPath, onSearch, children }: Props) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { theme, toggleTheme } = useTheme();
  const appConfig = useAppConfig();

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
        <h1 class={styles.title}>
          <a href="/" style={{ color: "inherit", textDecoration: "none" }}>
            {appConfig.title}
          </a>
        </h1>
        <nav class={styles.nav}>
          <a href="/graph" class={styles.navLink}>
            Graph
          </a>
        </nav>
        {onSearch && <SearchBar onSearch={onSearch} />}
        <ThemeToggle theme={theme} onToggle={toggleTheme} />
      </header>
      <div class={styles.body}>
        <div
          class={`${styles.sidebarWrapper} ${sidebarOpen ? styles.open : ""}`}
        >
          <Sidebar currentPath={currentPath} />
        </div>
        {sidebarOpen && (
          <div class={styles.overlay} onClick={() => setSidebarOpen(false)} />
        )}
        <main class={styles.main}>{children}</main>
      </div>
    </div>
  );
}
