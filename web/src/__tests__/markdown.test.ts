import { describe, it, expect } from "vitest";
import { renderMarkdown } from "../lib/markdown";

describe("renderMarkdown", () => {
  it("converts headings to HTML", () => {
    const html = renderMarkdown("# Hello World");
    expect(html).toContain("<h1");
    expect(html).toContain("Hello World");
    expect(html).toContain('id="hello-world"');
  });

  it("renders code blocks with syntax highlighting", () => {
    const md = '```go\nfunc main() {}\n```';
    const html = renderMarkdown(md);
    expect(html).toContain("hljs");
    expect(html).toContain("func");
  });

  it("renders GFM tables", () => {
    const md = "| A | B |\n|---|---|\n| 1 | 2 |";
    const html = renderMarkdown(md);
    expect(html).toContain("<table>");
    expect(html).toContain("<th>");
  });

  it("renders inline code", () => {
    const html = renderMarkdown("Use `go build` to compile.");
    expect(html).toContain("<code>");
    expect(html).toContain("go build");
  });

  it("renders lists", () => {
    const html = renderMarkdown("- item 1\n- item 2\n");
    expect(html).toContain("<ul>");
    expect(html).toContain("<li>");
  });

  it("resolves relative image paths with docPath", () => {
    const html = renderMarkdown("![alt](./images/diagram.png)", "docs/guide.md");
    expect(html).toContain("/api/v1/raw/docs/images/diagram.png");
  });

  it("preserves absolute image URLs", () => {
    const html = renderMarkdown("![alt](https://example.com/img.png)", "docs/guide.md");
    expect(html).toContain("https://example.com/img.png");
  });
});
