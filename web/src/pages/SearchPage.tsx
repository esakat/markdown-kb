import { useState, useEffect, useCallback } from "preact/hooks";
import { route } from "preact-router";
import { searchDocuments, listDocuments } from "../api/client";
import type { SearchResult } from "../types/api";
import { SearchResults } from "../components/Search/SearchResults";
import { FacetPanel } from "../components/Search/FacetPanel";

interface Props {
  path?: string;
  q?: string;
  status?: string;
  tag?: string;
  page?: string;
}

export function SearchPage({
  q = "",
  status = "",
  tag = "",
  page: pageStr = "1",
}: Props) {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  const query = q;
  const currentPage = Math.max(1, parseInt(pageStr, 10) || 1);
  const limit = 20;
  const filters: Record<string, string> = {};
  if (status) filters.status = status;
  if (tag) filters.tag = tag;

  const buildUrl = useCallback(
    (params: { q?: string; status?: string; tag?: string; page?: number }) => {
      const p = new URLSearchParams();
      const newQ = params.q ?? q;
      const newStatus = params.status ?? status;
      const newTag = params.tag ?? tag;
      const newPage = params.page ?? currentPage;

      if (newQ) p.set("q", newQ);
      if (newStatus) p.set("status", newStatus);
      if (newTag) p.set("tag", newTag);
      if (newPage > 1) p.set("page", String(newPage));

      const qs = p.toString();
      return `/search${qs ? `?${qs}` : ""}`;
    },
    [q, status, tag, currentPage],
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const apiFilters: Record<string, string> = {};
    if (status) apiFilters.status = status;
    if (tag) apiFilters.tag = tag;

    const promise = query
      ? searchDocuments(query, currentPage, limit, apiFilters)
      : Object.keys(apiFilters).length > 0
        ? listDocuments(currentPage, limit, apiFilters).then((res) => ({
            ...res,
            data: res.data.map((d) => ({
              path: d.path,
              title: d.title,
              snippet: "",
              score: 0,
              meta: d.meta,
            })),
          }))
        : Promise.resolve({
            data: [] as SearchResult[],
            total: 0,
            page: currentPage,
            limit,
          });

    promise
      .then((res) => {
        if (!cancelled) {
          setResults(res.data);
          setTotal(res.total);
          setLoading(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setResults([]);
          setTotal(0);
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [query, status, tag, currentPage]);

  const handleFilterChange = (newFilters: Record<string, string>) => {
    route(
      buildUrl({
        status: newFilters.status || "",
        tag: newFilters.tag || "",
        page: 1,
      }),
    );
  };

  const handlePageChange = (newPage: number) => {
    route(buildUrl({ page: newPage }));
  };

  return (
    <div style={{ display: "flex", gap: "0", height: "100%" }}>
      <FacetPanel filters={filters} onFilterChange={handleFilterChange} />
      <div style={{ flex: 1, padding: "0 24px", overflow: "auto" }}>
        <SearchResults
          results={results}
          total={total}
          page={currentPage}
          limit={limit}
          loading={loading}
          query={query}
          onPageChange={handlePageChange}
        />
      </div>
    </div>
  );
}
