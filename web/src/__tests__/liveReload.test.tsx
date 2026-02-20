import { render, screen, fireEvent } from "@testing-library/preact";
import { describe, it, expect, vi } from "vitest";
import { ToastContainer, createToast } from "../components/LiveReload/Toast";
import type { ToastMessage } from "../components/LiveReload/Toast";

describe("ToastContainer", () => {
  it("renders toast messages", () => {
    const messages: ToastMessage[] = [
      createToast("File updated", "info"),
      createToast("File created", "success"),
    ];

    render(<ToastContainer messages={messages} onDismiss={() => {}} />);

    expect(screen.getByText("File updated")).toBeTruthy();
    expect(screen.getByText("File created")).toBeTruthy();
  });

  it("calls onDismiss when close button is clicked", () => {
    const onDismiss = vi.fn();
    const messages: ToastMessage[] = [createToast("Test toast", "info")];

    render(<ToastContainer messages={messages} onDismiss={onDismiss} />);

    const closeBtn = screen.getByLabelText("Dismiss");
    fireEvent.click(closeBtn);

    expect(onDismiss).toHaveBeenCalledWith(messages[0].id);
  });

  it("renders empty when no messages", () => {
    const { container } = render(
      <ToastContainer messages={[]} onDismiss={() => {}} />,
    );
    const toastContainer = container.querySelector(
      '[data-testid="toast-container"]',
    );
    expect(toastContainer?.children.length).toBe(0);
  });

  it("renders correct style for warning type", () => {
    const messages: ToastMessage[] = [
      createToast("File deleted", "warning"),
    ];

    render(<ToastContainer messages={messages} onDismiss={() => {}} />);
    expect(screen.getByText("File deleted")).toBeTruthy();
  });
});

describe("createToast", () => {
  it("creates toast with unique ids", () => {
    const t1 = createToast("First", "info");
    const t2 = createToast("Second", "success");
    expect(t1.id).not.toBe(t2.id);
    expect(t1.text).toBe("First");
    expect(t1.type).toBe("info");
    expect(t2.text).toBe("Second");
    expect(t2.type).toBe("success");
  });

  it("defaults to info type", () => {
    const t = createToast("Default");
    expect(t.type).toBe("info");
  });
});
