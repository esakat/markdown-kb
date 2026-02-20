import { extractToc } from "../../lib/toc";
import styles from "./TableOfContents.module.css";

interface Props {
  markdown: string;
}

export function TableOfContents({ markdown }: Props) {
  const entries = extractToc(markdown);

  if (entries.length < 2) return null;

  const minLevel = Math.min(...entries.map((e) => e.level));

  return (
    <nav class={styles.toc} data-testid="toc">
      <h3 class={styles.title}>Contents</h3>
      <ul class={styles.list}>
        {entries.map((entry) => (
          <li
            key={entry.id}
            class={styles.item}
            style={{ paddingLeft: `${(entry.level - minLevel) * 16}px` }}
          >
            <a
              href={`#${entry.id}`}
              class={styles.link}
              onClick={(e) => {
                e.preventDefault();
                document.getElementById(entry.id)?.scrollIntoView({ behavior: "smooth" });
              }}
            >
              {entry.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}
