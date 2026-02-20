import { render, screen } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { SearchResults } from "../components/Search/SearchResults";
import type { SearchResult } from "../types/api";

const sampleResults: SearchResult[] = [
  {
    path: "guide.md",
    title: "Go Guide",
    snippet: "Learn <b>Go</b> programming language.",
    score: -1.5,
    meta: { status: "published", tags: ["go", "tutorial"] },
  },
  {
    path: "api.md",
    title: "API Reference",
    snippet: "REST <b>API</b> documentation.",
    score: -2.0,
    meta: { status: "draft", tags: ["go", "api"] },
  },
];

describe("SearchResults", () => {
  it("renders loading skeleton", () => {
    render(
      <SearchResults
        results={[]}
        total={0}
        page={1}
        limit={20}
        loading={true}
        query="test"
        onPageChange={() => {}}
      />
    );
    expect(screen.getByTestId("search-loading")).toBeTruthy();
  });

  it("renders empty state when no query", () => {
    render(
      <SearchResults
        results={[]}
        total={0}
        page={1}
        limit={20}
        loading={false}
        query=""
        onPageChange={() => {}}
      />
    );
    expect(screen.getByText("Search your documents")).toBeTruthy();
  });

  it("renders no results message", () => {
    render(
      <SearchResults
        results={[]}
        total={0}
        page={1}
        limit={20}
        loading={false}
        query="nonexistent"
        onPageChange={() => {}}
      />
    );
    expect(screen.getByTestId("search-empty")).toBeTruthy();
    expect(screen.getByText("No results found")).toBeTruthy();
  });

  it("renders search result items with titles", () => {
    render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="Go"
        onPageChange={() => {}}
      />
    );
    expect(screen.getByText("Go Guide")).toBeTruthy();
    expect(screen.getByText("API Reference")).toBeTruthy();
  });

  it("renders result count", () => {
    render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="Go"
        onPageChange={() => {}}
      />
    );
    expect(screen.getByText(/2 results/)).toBeTruthy();
  });

  it("renders snippets with HTML highlighting", () => {
    const { container } = render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="Go"
        onPageChange={() => {}}
      />
    );
    const bold = container.querySelector("b");
    expect(bold?.textContent).toBe("Go");
  });

  it("renders pagination when multiple pages", () => {
    const onPageChange = vi.fn();
    render(
      <SearchResults
        results={sampleResults}
        total={50}
        page={1}
        limit={20}
        loading={false}
        query="test"
        onPageChange={onPageChange}
      />
    );
    expect(screen.getByText("Page 1 of 3")).toBeTruthy();
    expect(screen.getByText("Previous")).toBeTruthy();
    expect(screen.getByText("Next")).toBeTruthy();
  });

  it("does not render pagination for single page", () => {
    render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="test"
        onPageChange={() => {}}
      />
    );
    expect(screen.queryByText("Previous")).toBeNull();
  });

  it("renders tags for results", () => {
    render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="Go"
        onPageChange={() => {}}
      />
    );
    expect(screen.getByText("tutorial")).toBeTruthy();
    expect(screen.getByText("api")).toBeTruthy();
  });

  it("result items link to document pages", () => {
    const { container } = render(
      <SearchResults
        results={sampleResults}
        total={2}
        page={1}
        limit={20}
        loading={false}
        query="Go"
        onPageChange={() => {}}
      />
    );
    const links = container.querySelectorAll("a[data-testid='search-result-item']");
    expect(links).toHaveLength(2);
    expect((links[0] as HTMLAnchorElement).getAttribute("href")).toBe("/docs/guide.md");
    expect((links[1] as HTMLAnchorElement).getAttribute("href")).toBe("/docs/api.md");
  });
});
