import { Marked } from "marked";
import { markedHighlight } from "marked-highlight";
import hljs from "highlight.js/lib/core";
import go from "highlight.js/lib/languages/go";
import typescript from "highlight.js/lib/languages/typescript";
import javascript from "highlight.js/lib/languages/javascript";
import python from "highlight.js/lib/languages/python";
import bash from "highlight.js/lib/languages/bash";
import yaml from "highlight.js/lib/languages/yaml";
import json from "highlight.js/lib/languages/json";
import sql from "highlight.js/lib/languages/sql";
import css from "highlight.js/lib/languages/css";
import xml from "highlight.js/lib/languages/xml";
import markdown from "highlight.js/lib/languages/markdown";

hljs.registerLanguage("go", go);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("python", python);
hljs.registerLanguage("bash", bash);
hljs.registerLanguage("shell", bash);
hljs.registerLanguage("yaml", yaml);
hljs.registerLanguage("json", json);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("css", css);
hljs.registerLanguage("html", xml);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("markdown", markdown);

const marked = new Marked(
  markedHighlight({
    langPrefix: "hljs language-",
    highlight(code: string, lang: string) {
      if (lang && hljs.getLanguage(lang)) {
        return hljs.highlight(code, { language: lang }).value;
      }
      return code;
    },
  }),
);

// Custom heading renderer to add id attributes for ToC linking
const renderer = {
  heading({ text, depth }: { text: string; depth: number }) {
    const id = text
      .toLowerCase()
      .replace(/<[^>]*>/g, "")
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-")
      .trim();
    return `<h${depth} id="${id}">${text}</h${depth}>`;
  },
};

marked.use({ renderer });

export function renderMarkdown(source: string, docPath?: string): string {
  if (docPath) {
    const dir = docPath.substring(0, docPath.lastIndexOf("/") + 1);
    const imgRenderer = {
      image({
        href,
        title,
        text,
      }: {
        href: string;
        title: string | null;
        text: string;
      }) {
        if (href && !href.startsWith("http") && !href.startsWith("/")) {
          // Strip leading ./ prefix
          const cleanHref = href.replace(/^\.\//, "");
          href = `/api/v1/raw/${dir}${cleanHref}`;
        }
        const titleAttr = title ? ` title="${title}"` : "";
        return `<img src="${href}" alt="${text}"${titleAttr} />`;
      },
    };
    const withImg = new Marked(
      markedHighlight({
        langPrefix: "hljs language-",
        highlight(code: string, lang: string) {
          if (lang && hljs.getLanguage(lang)) {
            return hljs.highlight(code, { language: lang }).value;
          }
          return code;
        },
      }),
    );
    withImg.use({ renderer: { ...renderer, ...imgRenderer } });
    return withImg.parse(source) as string;
  }

  return marked.parse(source) as string;
}
