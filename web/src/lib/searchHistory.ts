const STORAGE_KEY = "markdown-kb:search-history";
const MAX_ITEMS = 10;

export function getSearchHistory(): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export function addSearchHistory(query: string): void {
  const trimmed = query.trim();
  if (!trimmed) return;

  const history = getSearchHistory().filter((item) => item !== trimmed);
  history.unshift(trimmed);

  localStorage.setItem(STORAGE_KEY, JSON.stringify(history.slice(0, MAX_ITEMS)));
}

export function clearSearchHistory(): void {
  localStorage.removeItem(STORAGE_KEY);
}
