import { render, screen } from "@testing-library/preact";
import { describe, it, expect } from "vitest";
import { App } from "../app";

describe("App", () => {
  it("renders the Markdown KB heading", () => {
    render(<App />);
    expect(screen.getByText("Markdown KB")).toBeTruthy();
  });

  it("renders welcome message on home route", () => {
    render(<App />);
    expect(screen.getByText("Welcome")).toBeTruthy();
  });
});
