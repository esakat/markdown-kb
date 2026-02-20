import { useState, useRef, useCallback, useEffect } from "preact/hooks";
import { getSearchHistory, addSearchHistory, clearSearchHistory } from "../../lib/searchHistory";
import styles from "./SearchBar.module.css";

interface Props {
  initialQuery?: string;
  onSearch: (query: string) => void;
}

export function SearchBar({ initialQuery = "", onSearch }: Props) {
  const [value, setValue] = useState(initialQuery);
  const [showHistory, setShowHistory] = useState(false);
  const [history, setHistory] = useState<string[]>([]);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setValue(initialQuery);
  }, [initialQuery]);

  const debouncedSearch = useCallback(
    (query: string) => {
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        if (query.trim()) {
          addSearchHistory(query.trim());
        }
        onSearch(query.trim());
      }, 300);
    },
    [onSearch]
  );

  const handleInput = (e: Event) => {
    const newValue = (e.target as HTMLInputElement).value;
    setValue(newValue);
    debouncedSearch(newValue);
  };

  const handleClear = () => {
    setValue("");
    if (timerRef.current) clearTimeout(timerRef.current);
    onSearch("");
  };

  const handleFocus = () => {
    if (!value.trim()) {
      setHistory(getSearchHistory());
      setShowHistory(true);
    }
  };

  const handleHistoryClick = (query: string) => {
    setValue(query);
    setShowHistory(false);
    if (timerRef.current) clearTimeout(timerRef.current);
    addSearchHistory(query);
    onSearch(query);
  };

  const handleClearHistory = () => {
    clearSearchHistory();
    setHistory([]);
    setShowHistory(false);
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      setShowHistory(false);
    }
    if (e.key === "Enter") {
      if (timerRef.current) clearTimeout(timerRef.current);
      const trimmed = value.trim();
      if (trimmed) addSearchHistory(trimmed);
      onSearch(trimmed);
      setShowHistory(false);
    }
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setShowHistory(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div class={styles.wrapper} ref={wrapperRef}>
      <div class={styles.inputWrapper}>
        <span class={styles.icon} aria-hidden="true">&#128269;</span>
        <input
          class={styles.input}
          type="text"
          placeholder="Search documents..."
          value={value}
          onInput={handleInput}
          onFocus={handleFocus}
          onKeyDown={handleKeyDown}
          data-testid="search-input"
        />
        {value && (
          <button
            class={styles.clearBtn}
            onClick={handleClear}
            aria-label="Clear search"
            data-testid="search-clear"
          >
            &#x2715;
          </button>
        )}
      </div>
      {showHistory && history.length > 0 && (
        <div class={styles.dropdown} data-testid="search-history">
          <div class={styles.dropdownHeader}>
            <span>Recent searches</span>
            <button class={styles.clearHistoryBtn} onClick={handleClearHistory}>
              Clear
            </button>
          </div>
          {history.map((item) => (
            <button
              key={item}
              class={styles.historyItem}
              onClick={() => handleHistoryClick(item)}
            >
              {item}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
