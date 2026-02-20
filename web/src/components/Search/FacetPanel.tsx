import { useState, useEffect } from "preact/hooks";
import { getMetadataFields, listTags } from "../../api/client";
import type { TagCount } from "../../types/api";
import styles from "./FacetPanel.module.css";

interface Props {
  filters: Record<string, string>;
  onFilterChange: (filters: Record<string, string>) => void;
}

export function FacetPanel({ filters, onFilterChange }: Props) {
  const [statusValues, setStatusValues] = useState<string[]>([]);
  const [tags, setTags] = useState<TagCount[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    Promise.all([getMetadataFields(), listTags()])
      .then(([fieldsRes, tagsRes]) => {
        if (cancelled) return;
        const statusField = fieldsRes.data.find((f) => f.name === "status");
        if (statusField) {
          setStatusValues(statusField.values.sort());
        }
        setTags(tagsRes.data.sort((a, b) => b.count - a.count));
        setLoading(false);
      })
      .catch(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const activeFilterEntries = Object.entries(filters).filter(([, v]) => v);
  const hasFilters = activeFilterEntries.length > 0;

  const toggleStatus = (status: string) => {
    const current = filters.status || "";
    const next = current === status ? "" : status;
    onFilterChange({ ...filters, status: next });
  };

  const toggleTag = (tag: string) => {
    const current = filters.tag || "";
    const next = current === tag ? "" : tag;
    onFilterChange({ ...filters, tag: next });
  };

  const removeFilter = (key: string) => {
    onFilterChange({ ...filters, [key]: "" });
  };

  const clearAll = () => {
    onFilterChange({});
  };

  if (loading) {
    return (
      <aside class={styles.panel}>
        <div class={styles.loading}>Loading filters...</div>
      </aside>
    );
  }

  return (
    <aside class={styles.panel} data-testid="facet-panel">
      {hasFilters && (
        <>
          <div class={styles.activeFilters}>
            {activeFilterEntries.map(([key, val]) => (
              <button
                key={key}
                class={styles.chip}
                onClick={() => removeFilter(key)}
                data-testid="active-filter-chip"
              >
                {key}: {val}
                <span class={styles.chipRemove}>&#x2715;</span>
              </button>
            ))}
          </div>
          <button class={styles.clearAll} onClick={clearAll}>
            Clear all filters
          </button>
        </>
      )}

      {statusValues.length > 0 && (
        <div class={styles.section}>
          <div class={styles.sectionTitle}>Status</div>
          {statusValues.map((status) => (
            <label key={status} class={styles.checkboxItem}>
              <input
                type="checkbox"
                checked={filters.status === status}
                onChange={() => toggleStatus(status)}
              />
              {status}
            </label>
          ))}
        </div>
      )}

      {tags.length > 0 && (
        <div class={styles.section}>
          <div class={styles.sectionTitle}>Tags</div>
          <div class={styles.tagList}>
            {tags.map((tc) => (
              <button
                key={tc.tag}
                class={`${styles.tagBtn} ${filters.tag === tc.tag ? styles.tagBtnActive : ""}`}
                onClick={() => toggleTag(tc.tag)}
              >
                {tc.tag}
                <span class={styles.tagCount}>{tc.count}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </aside>
  );
}
