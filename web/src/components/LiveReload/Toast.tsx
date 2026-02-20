import { useEffect } from "preact/hooks";
import styles from "./Toast.module.css";

interface ToastMessage {
  id: number;
  text: string;
  type: "info" | "success" | "warning";
}

let nextId = 0;

interface Props {
  messages: ToastMessage[];
  onDismiss: (id: number) => void;
}

export function ToastContainer({ messages, onDismiss }: Props) {
  return (
    <div class={styles.container} data-testid="toast-container">
      {messages.map((msg) => (
        <ToastItem key={msg.id} message={msg} onDismiss={onDismiss} />
      ))}
    </div>
  );
}

function ToastItem({
  message,
  onDismiss,
}: {
  message: ToastMessage;
  onDismiss: (id: number) => void;
}) {
  useEffect(() => {
    const timer = setTimeout(() => onDismiss(message.id), 4000);
    return () => clearTimeout(timer);
  }, [message.id, onDismiss]);

  return (
    <div class={`${styles.toast} ${styles[message.type]}`}>
      <span>{message.text}</span>
      <button
        class={styles.close}
        onClick={() => onDismiss(message.id)}
        aria-label="Dismiss"
      >
        ×
      </button>
    </div>
  );
}

export function createToast(
  text: string,
  type: "info" | "success" | "warning" = "info",
): ToastMessage {
  return { id: nextId++, text, type };
}

export type { ToastMessage };
