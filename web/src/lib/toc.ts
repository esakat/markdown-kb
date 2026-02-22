export interface TocEntry {
  id: string;
  text: string;
  level: number;
}

export function extractToc(markdown: string): TocEntry[] {
  const headingRegex = /^(#{1,6})\s+(.+)$/gm;
  const entries: TocEntry[] = [];
  let match;

  while ((match = headingRegex.exec(markdown)) !== null) {
    const level = match[1].length;
    const text = match[2].trim();
    const id = text
      .toLowerCase()
      .replace(/<[^>]*>/g, "")
      .replace(/[^\w\s\u3000-\u9fff\uff00-\uffef-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-")
      .trim();
    entries.push({ id, text, level });
  }

  return entries;
}
