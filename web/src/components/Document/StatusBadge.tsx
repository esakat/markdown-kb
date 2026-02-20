import styles from "./StatusBadge.module.css";

interface Props {
  status: string;
}

const statusColors: Record<string, string> = {
  published: styles.published,
  draft: styles.draft,
  archived: styles.archived,
};

export function StatusBadge({ status }: Props) {
  const colorClass = statusColors[status] || styles.default;
  return <span class={`${styles.badge} ${colorClass}`}>{status}</span>;
}
