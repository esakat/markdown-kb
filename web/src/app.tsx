import Router, { route as _route } from "preact-router";
import { useState } from "preact/hooks";
import { Layout } from "./components/Layout";
import { Home } from "./pages/Home";
import { DocumentPage } from "./pages/DocumentPage";

export function App() {
  const [currentPath, setCurrentPath] = useState<string | undefined>(undefined);

  const handleRoute = (e: { url: string }) => {
    const match = e.url.match(/^\/docs\/(.+)$/);
    setCurrentPath(match ? match[1] : undefined);
  };

  return (
    <Layout currentPath={currentPath}>
      <Router onChange={handleRoute}>
        <Home path="/" />
        <DocumentPage path="/docs/:docPath*" />
      </Router>
    </Layout>
  );
}
