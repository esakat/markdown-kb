import {
  Marked,
  type TokenizerExtension,
  type RendererExtension,
  type Tokens,
} from "marked";
import { markedHighlight } from "marked-highlight";
import hljs from "highlight.js/lib/core";

// --- Register highlight.js languages ---
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
import diff from "highlight.js/lib/languages/diff";
import dockerfile from "highlight.js/lib/languages/dockerfile";
import rust from "highlight.js/lib/languages/rust";
import java from "highlight.js/lib/languages/java";
import c from "highlight.js/lib/languages/c";
import cpp from "highlight.js/lib/languages/cpp";
import ruby from "highlight.js/lib/languages/ruby";
import php from "highlight.js/lib/languages/php";
import swift from "highlight.js/lib/languages/swift";
import kotlin from "highlight.js/lib/languages/kotlin";
import plaintext from "highlight.js/lib/languages/plaintext";

hljs.registerLanguage("go", go);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("ts", typescript);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("js", javascript);
hljs.registerLanguage("python", python);
hljs.registerLanguage("bash", bash);
hljs.registerLanguage("shell", bash);
hljs.registerLanguage("sh", bash);
hljs.registerLanguage("yaml", yaml);
hljs.registerLanguage("yml", yaml);
hljs.registerLanguage("json", json);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("css", css);
hljs.registerLanguage("html", xml);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("markdown", markdown);
hljs.registerLanguage("md", markdown);
hljs.registerLanguage("diff", diff);
hljs.registerLanguage("dockerfile", dockerfile);
hljs.registerLanguage("docker", dockerfile);
hljs.registerLanguage("rust", rust);
hljs.registerLanguage("java", java);
hljs.registerLanguage("c", c);
hljs.registerLanguage("cpp", cpp);
hljs.registerLanguage("ruby", ruby);
hljs.registerLanguage("rb", ruby);
hljs.registerLanguage("php", php);
hljs.registerLanguage("swift", swift);
hljs.registerLanguage("kotlin", kotlin);
hljs.registerLanguage("plaintext", plaintext);
hljs.registerLanguage("text", plaintext);

// --- Highlight function ---
function highlight(code: string, lang: string): string {
  if (lang && hljs.getLanguage(lang)) {
    return hljs.highlight(code, { language: lang }).value;
  }
  return code;
}

// --- Mermaid code block extension ---
// Renders mermaid code blocks as <pre class="mermaid"> for client-side rendering
const mermaidExtension: TokenizerExtension & RendererExtension = {
  name: "mermaidBlock",
  level: "block",
  start(src: string) {
    return src.indexOf("```mermaid");
  },
  tokenizer(src: string) {
    const match = src.match(/^```mermaid\n([\s\S]*?)```/);
    if (match) {
      return {
        type: "mermaidBlock",
        raw: match[0],
        text: match[1].trim(),
      };
    }
    return undefined;
  },
  renderer(token: Tokens.Generic) {
    const escaped = (token.text as string)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
    return `<pre class="mermaid">${escaped}</pre>`;
  },
};

// --- Wiki-link extension: [[path]] or [[path|label]] ---
const wikiLinkExtension: TokenizerExtension & RendererExtension = {
  name: "wikiLink",
  level: "inline",
  start(src: string) {
    return src.indexOf("[[");
  },
  tokenizer(src: string) {
    const match = src.match(/^\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/);
    if (match) {
      const rawPath = match[1].trim();
      const label = match[2]?.trim();
      // Ensure .md extension
      const path = rawPath.endsWith(".md") ? rawPath : rawPath + ".md";
      return {
        type: "wikiLink",
        raw: match[0],
        path,
        label: label || rawPath.split("/").pop() || rawPath,
      };
    }
    return undefined;
  },
  renderer(token: Tokens.Generic) {
    const path = token.path as string;
    const label = token.label as string;
    return `<a href="/docs/${path}" class="wiki-link" title="${path}">${label}</a>`;
  },
};

// --- :::message container directive ---
// Supports :::message, :::warning, :::info, :::tip, :::danger, :::details
// Pre-processes source markdown before marked parsing.
// Wraps content in raw HTML div so marked passes it through while still
// parsing the inner markdown (blank lines around body enable this).
function preprocessContainers(source: string): string {
  return source.replace(
    /^:::(\w+)(?:[ \t]+(.+))?\n([\s\S]*?)\n:::/gm,
    (_match, variant: string, title: string | undefined, body: string) => {
      if (variant === "details") {
        const summary = title || "";
        return `<details class="container-details"><summary>${summary}</summary>\n\n${body}\n\n</details>`;
      }
      return `<div class="container-${variant}">\n\n${body}\n\n</div>`;
    },
  );
}

// --- Custom heading renderer with id ---
const headingRenderer = {
  heading({ text, depth }: { text: string; depth: number }) {
    const id = text
      .toLowerCase()
      .replace(/<[^>]*>/g, "")
      .replace(/[^\w\s\u3000-\u9fff\uff00-\uffef-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-")
      .trim();
    return `<h${depth} id="${id}">${text}</h${depth}>`;
  },
};

// --- Build Marked instance ---
function buildMarked(docPath?: string) {
  const m = new Marked(
    markedHighlight({
      langPrefix: "hljs language-",
      highlight,
    }),
  );

  m.use({
    extensions: [mermaidExtension, wikiLinkExtension],
  });

  const renderers: Record<string, unknown> = { ...headingRenderer };

  // Image path resolution for relative images
  if (docPath) {
    const dir = docPath.substring(0, docPath.lastIndexOf("/") + 1);
    renderers.image = ({
      href,
      title,
      text,
    }: {
      href: string;
      title: string | null;
      text: string;
    }) => {
      if (href && !href.startsWith("http") && !href.startsWith("/")) {
        const cleanHref = href.replace(/^\.\//, "");
        href = `/api/v1/raw/${dir}${cleanHref}`;
      }
      const titleAttr = title ? ` title="${title}"` : "";
      return `<img src="${href}" alt="${text}"${titleAttr} />`;
    };
  }

  m.use({ renderer: renderers });
  return m;
}

export function renderMarkdown(source: string, docPath?: string): string {
  const m = buildMarked(docPath);
  const preprocessed = preprocessContainers(source);
  return m.parse(preprocessed) as string;
}
