import Router from "preact-router";
import { useState, useCallback } from "preact/hooks";
import { route } from "preact-router";
import { Layout } from "./components/Layout";
import { Home } from "./pages/Home";
import { DocumentPage } from "./pages/DocumentPage";
import { SearchPage } from "./pages/SearchPage";
import { GraphPage } from "./pages/GraphPage";
import { useWebSocket } from "./hooks/useWebSocket";
import type { WSEvent } from "./hooks/useWebSocket";
import { ToastContainer, createToast } from "./components/LiveReload/Toast";
import type { ToastMessage } from "./components/LiveReload/Toast";

export function App() {
  const [currentPath, setCurrentPath] = useState<string | undefined>(undefined);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const [refreshKey, setRefreshKey] = useState(0);

  const addToast = useCallback(
    (text: string, type: "info" | "success" | "warning" = "info") => {
      setToasts((prev) => [...prev.slice(-4), createToast(text, type)]);
    },
    [],
  );

  const dismissToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const handleWSEvent = useCallback(
    (event: WSEvent) => {
      const fileName = event.path.split("/").pop() || event.path;
      switch (event.type) {
        case "created":
          addToast(`New: ${fileName}`, "success");
          break;
        case "updated":
          addToast(`Updated: ${fileName}`, "info");
          break;
        case "deleted":
          addToast(`Deleted: ${fileName}`, "warning");
          break;
      }
      // Trigger re-render for data-dependent components
      setRefreshKey((k) => k + 1);
    },
    [addToast],
  );

  useWebSocket({ onEvent: handleWSEvent });

  const handleRoute = (e: { url: string }) => {
    const match = e.url.match(/^\/docs\/(.+)$/);
    setCurrentPath(match ? match[1] : undefined);
  };

  const handleSearch = useCallback((query: string) => {
    if (query) {
      route(`/search?q=${encodeURIComponent(query)}`);
    } else {
      route("/search");
    }
  }, []);

  return (
    <Layout currentPath={currentPath} onSearch={handleSearch}>
      <Router onChange={handleRoute}>
        <Home path="/" key={`home-${refreshKey}`} />
        <SearchPage path="/search" />
        <GraphPage path="/graph" key={`graph-${refreshKey}`} />
        <DocumentPage
          path="/docs/:docPath*"
          key={`doc-${currentPath}-${refreshKey}`}
        />
      </Router>
      <ToastContainer messages={toasts} onDismiss={dismissToast} />
    </Layout>
  );
}
