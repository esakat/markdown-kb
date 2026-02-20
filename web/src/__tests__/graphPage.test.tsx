import { render, screen } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

import { TagCloud } from "../components/Graph/TagCloud";

describe("TagCloud", () => {
  it("renders tags with correct counts", () => {
    const tags = [
      { tag: "go", count: 5 },
      { tag: "tutorial", count: 2 },
    ];

    render(<TagCloud tags={tags} />);

    expect(screen.getByText("go")).toBeTruthy();
    expect(screen.getByText("tutorial")).toBeTruthy();
    expect(screen.getByText("5")).toBeTruthy();
    expect(screen.getByText("2")).toBeTruthy();
  });

  it("renders empty message when no tags", () => {
    render(<TagCloud tags={[]} />);
    expect(screen.getByText("No tags found.")).toBeTruthy();
  });

  it("highlights selected tag", () => {
    const tags = [
      { tag: "go", count: 3 },
      { tag: "api", count: 1 },
    ];

    render(<TagCloud tags={tags} selectedTag="go" />);

    const goBtn = screen.getByText("go").closest("button");
    expect(goBtn?.className).toContain("active");
  });

  it("calls onTagClick with tag name", () => {
    const onTagClick = vi.fn();
    const tags = [{ tag: "go", count: 3 }];

    render(<TagCloud tags={tags} onTagClick={onTagClick} />);

    screen.getByText("go").closest("button")?.click();
    expect(onTagClick).toHaveBeenCalledWith("go");
  });

  it("calls onTagClick with undefined when deselecting", () => {
    const onTagClick = vi.fn();
    const tags = [{ tag: "go", count: 3 }];

    render(<TagCloud tags={tags} selectedTag="go" onTagClick={onTagClick} />);

    screen.getByText("go").closest("button")?.click();
    expect(onTagClick).toHaveBeenCalledWith(undefined);
  });
});
