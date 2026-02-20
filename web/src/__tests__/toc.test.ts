import { describe, it, expect } from "vitest";
import { extractToc } from "../lib/toc";

describe("extractToc", () => {
  it("extracts headings from markdown", () => {
    const md =
      "# Title\n\nSome text.\n\n## Section 1\n\nMore text.\n\n### Subsection\n";
    const entries = extractToc(md);

    expect(entries).toHaveLength(3);
    expect(entries[0]).toEqual({ id: "title", text: "Title", level: 1 });
    expect(entries[1]).toEqual({
      id: "section-1",
      text: "Section 1",
      level: 2,
    });
    expect(entries[2]).toEqual({
      id: "subsection",
      text: "Subsection",
      level: 3,
    });
  });

  it("returns empty array for no headings", () => {
    const entries = extractToc("Just plain text.\n\nNo headings here.");
    expect(entries).toHaveLength(0);
  });

  it("handles special characters in headings", () => {
    const entries = extractToc("## API & Configuration (v2)");
    expect(entries).toHaveLength(1);
    expect(entries[0].id).toBe("api-configuration-v2");
  });

  it("handles Japanese headings", () => {
    const entries = extractToc("## 日本語ガイド");
    expect(entries).toHaveLength(1);
    expect(entries[0].text).toBe("日本語ガイド");
  });
});
