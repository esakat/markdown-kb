import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock fetch globally
const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

import { CommitHistory } from "../components/Git/CommitHistory";
import { DiffView } from "../components/Git/DiffView";

describe("CommitHistory", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders collapsed by default", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () =>
        Promise.resolve({
          data: [
            {
              hash: "abc1234567890abcdef1234567890abcdef123456",
              author: "Test",
              date: "2026-01-01T00:00:00Z",
              message: "initial commit",
            },
          ],
        }),
    });

    render(<CommitHistory docPath="guide.md" />);

    expect(screen.getByText(/Commit History/)).toBeTruthy();
    // The list should not be visible initially
    expect(screen.queryByText("initial commit")).toBeNull();
  });

  it("shows commits when expanded", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () =>
        Promise.resolve({
          data: [
            {
              hash: "abc1234567890abcdef1234567890abcdef123456",
              author: "Test Author",
              date: "2026-01-15T10:30:00Z",
              message: "update: something important",
            },
            {
              hash: "def4567890abcdef1234567890abcdef123456789",
              author: "Test Author",
              date: "2026-01-01T00:00:00Z",
              message: "initial: first commit",
            },
          ],
        }),
    });

    render(<CommitHistory docPath="guide.md" />);

    // Wait for fetch to resolve
    await new Promise((r) => setTimeout(r, 10));

    // Click to expand
    fireEvent.click(screen.getByText(/Commit History/));

    // Now commits should be visible
    expect(screen.getByText("update: something important")).toBeTruthy();
    expect(screen.getByText("initial: first commit")).toBeTruthy();
    expect(screen.getAllByText("Test Author")).toHaveLength(2);
  });

  it("shows diff button for non-last commits", async () => {
    const onSelectDiff = vi.fn();
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () =>
        Promise.resolve({
          data: [
            {
              hash: "abc1234567890abcdef1234567890abcdef123456",
              author: "Test",
              date: "2026-01-15T00:00:00Z",
              message: "second",
            },
            {
              hash: "def4567890abcdef1234567890abcdef123456789",
              author: "Test",
              date: "2026-01-01T00:00:00Z",
              message: "first",
            },
          ],
        }),
    });

    render(<CommitHistory docPath="guide.md" onSelectDiff={onSelectDiff} />);

    await new Promise((r) => setTimeout(r, 10));
    fireEvent.click(screen.getByText(/Commit History/));

    const diffBtn = screen.getByText("diff");
    expect(diffBtn).toBeTruthy();
  });
});

describe("DiffView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders diff with added and removed lines", async () => {
    const diffOutput = `diff --git a/guide.md b/guide.md
index abc..def 100644
--- a/guide.md
+++ b/guide.md
@@ -1,3 +1,4 @@
 # Hello

-First version.
+Updated version.
+New line added.`;

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: diffOutput }),
    });

    const onClose = vi.fn();
    render(
      <DiffView
        docPath="guide.md"
        fromHash="abc1234"
        toHash="def5678"
        onClose={onClose}
      />,
    );

    await new Promise((r) => setTimeout(r, 10));

    expect(screen.getByTestId("diff-view")).toBeTruthy();
    expect(screen.getByText(/abc1234/)).toBeTruthy();
    expect(screen.getByText(/def5678/)).toBeTruthy();
  });

  it("calls onClose when close button is clicked", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: "" }),
    });

    const onClose = vi.fn();
    render(
      <DiffView
        docPath="guide.md"
        fromHash="abc"
        toHash="def"
        onClose={onClose}
      />,
    );

    await new Promise((r) => setTimeout(r, 10));

    fireEvent.click(screen.getByLabelText("Close diff"));
    expect(onClose).toHaveBeenCalled();
  });

  it("shows empty state when no diff", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: "" }),
    });

    render(
      <DiffView
        docPath="guide.md"
        fromHash="abc"
        toHash="def"
        onClose={() => {}}
      />,
    );

    await new Promise((r) => setTimeout(r, 10));

    expect(screen.getByText("No changes between these commits.")).toBeTruthy();
  });
});
