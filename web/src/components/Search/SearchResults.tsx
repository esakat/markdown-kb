import type { SearchResult } from "../../types/api";
import { TagBadge } from "../Document/TagBadge";
import { StatusBadge } from "../Document/StatusBadge";
import styles from "./SearchResults.module.css";

interface Props {
  results: SearchResult[];
  total: number;
  page: number;
  limit: number;
  loading: boolean;
  query: string;
  onPageChange: (page: number) => void;
}

export function SearchResults({
  results,
  total,
  page,
  limit,
  loading,
  query,
  onPageChange,
}: Props) {
  const totalPages = Math.ceil(total / limit);

  if (loading) {
    return (
      <div class={styles.loading} data-testid="search-loading">
        {[1, 2, 3].map((i) => (
          <div key={i} class={styles.skeleton} />
        ))}
      </div>
    );
  }

  if (!query) {
    return (
      <div class={styles.empty}>
        <div class={styles.emptyTitle}>Search your documents</div>
        <p>Type a query in the search bar to find documents.</p>
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div class={styles.empty} data-testid="search-empty">
        <div class={styles.emptyTitle}>No results found</div>
        <p>No documents match &ldquo;{query}&rdquo;. Try a different search term.</p>
      </div>
    );
  }

  return (
    <div class={styles.container}>
      <div class={styles.header}>
        <span>
          {total} result{total !== 1 ? "s" : ""} for &ldquo;{query}&rdquo;
        </span>
      </div>

      {results.map((result) => {
        const tags = extractTags(result.meta);
        const status = result.meta?.status as string | undefined;

        return (
          <a
            key={result.path}
            class={styles.resultItem}
            href={`/docs/${result.path}`}
            data-testid="search-result-item"
          >
            <div class={styles.resultTitle}>{result.title || result.path}</div>
            <div
              class={styles.resultSnippet}
              dangerouslySetInnerHTML={{ __html: result.snippet }}
            />
            <div class={styles.resultMeta}>
              <span class={styles.resultPath}>{result.path}</span>
              {status && <StatusBadge status={status} />}
              {tags.length > 0 && (
                <span class={styles.tags}>
                  {tags.map((tag) => (
                    <TagBadge key={tag} tag={tag} />
                  ))}
                </span>
              )}
            </div>
          </a>
        );
      })}

      {totalPages > 1 && (
        <div class={styles.pagination}>
          <button
            class={styles.pageBtn}
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
          >
            Previous
          </button>
          <span class={styles.pageInfo}>
            Page {page} of {totalPages}
          </span>
          <button
            class={styles.pageBtn}
            disabled={page >= totalPages}
            onClick={() => onPageChange(page + 1)}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

function extractTags(meta: Record<string, unknown>): string[] {
  const tags = meta?.tags;
  if (Array.isArray(tags)) return tags.filter((t): t is string => typeof t === "string");
  if (typeof tags === "string") return [tags];
  return [];
}
