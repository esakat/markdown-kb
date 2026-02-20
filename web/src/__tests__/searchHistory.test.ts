import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  getSearchHistory,
  addSearchHistory,
  clearSearchHistory,
} from "../lib/searchHistory";

// Mock localStorage with a simple in-memory implementation
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
})();

Object.defineProperty(globalThis, "localStorage", {
  value: localStorageMock,
  writable: true,
});

describe("searchHistory", () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  it("returns empty array when no history exists", () => {
    expect(getSearchHistory()).toEqual([]);
  });

  it("adds a query to history", () => {
    addSearchHistory("golang");
    expect(getSearchHistory()).toEqual(["golang"]);
  });

  it("prepends new queries (most recent first)", () => {
    addSearchHistory("first");
    addSearchHistory("second");
    expect(getSearchHistory()).toEqual(["second", "first"]);
  });

  it("deduplicates queries (moves to front)", () => {
    addSearchHistory("alpha");
    addSearchHistory("beta");
    addSearchHistory("alpha");
    expect(getSearchHistory()).toEqual(["alpha", "beta"]);
  });

  it("trims whitespace from queries", () => {
    addSearchHistory("  hello  ");
    expect(getSearchHistory()).toEqual(["hello"]);
  });

  it("ignores empty or whitespace-only queries", () => {
    addSearchHistory("");
    addSearchHistory("   ");
    expect(getSearchHistory()).toEqual([]);
  });

  it("limits history to 10 items", () => {
    for (let i = 0; i < 15; i++) {
      addSearchHistory(`query-${i}`);
    }
    const history = getSearchHistory();
    expect(history).toHaveLength(10);
    expect(history[0]).toBe("query-14");
  });

  it("clears all history", () => {
    addSearchHistory("test");
    clearSearchHistory();
    expect(getSearchHistory()).toEqual([]);
  });

  it("handles corrupted localStorage gracefully", () => {
    localStorageMock.setItem("markdown-kb:search-history", "not-json");
    expect(getSearchHistory()).toEqual([]);
  });
});
