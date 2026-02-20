import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { TreeView } from "../components/Sidebar/TreeView";
import type { TreeNode } from "../types/api";

const mockTree: TreeNode = {
  name: "",
  type: "dir",
  children: [
    {
      name: "docs",
      type: "dir",
      children: [
        { name: "api.md", type: "file", path: "docs/api.md", title: "API Reference" },
      ],
    },
    { name: "guide.md", type: "file", path: "guide.md", title: "Guide" },
  ],
};

describe("TreeView", () => {
  it("renders root children", () => {
    render(<TreeView tree={mockTree} />);
    expect(screen.getByTestId("tree-view")).toBeTruthy();
    expect(screen.getByTestId("tree-node-docs")).toBeTruthy();
    expect(screen.getByTestId("tree-node-guide.md")).toBeTruthy();
  });

  it("shows empty message when tree has no children", () => {
    const emptyTree: TreeNode = { name: "", type: "dir", children: [] };
    render(<TreeView tree={emptyTree} />);
    expect(screen.getByText("No documents found.")).toBeTruthy();
  });

  it("expands directory on click", () => {
    render(<TreeView tree={mockTree} />);

    const docsNode = screen.getByTestId("tree-node-docs");
    // Initially expanded (depth 0)
    expect(screen.getByTestId("tree-node-api.md")).toBeTruthy();

    // Click to collapse
    fireEvent.click(docsNode);
    expect(screen.queryByTestId("tree-node-api.md")).toBeNull();

    // Click to expand again
    fireEvent.click(docsNode);
    expect(screen.getByTestId("tree-node-api.md")).toBeTruthy();
  });

  it("displays title for file nodes", () => {
    render(<TreeView tree={mockTree} />);
    expect(screen.getByText("Guide")).toBeTruthy();
    expect(screen.getByText("API Reference")).toBeTruthy();
  });
});
