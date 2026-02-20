import styles from "./TagBadge.module.css";

interface Props {
  tag: string;
}

export function TagBadge({ tag }: Props) {
  return <span class={styles.badge}>{tag}</span>;
}
