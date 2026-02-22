export interface TreeNode {
  name: string;
  type: "dir" | "file";
  path?: string;
  title?: string;
  tags?: string[];
  children?: TreeNode[];
}

export interface TagIcon {
  tag: string;
  emoji: string;
}

export interface DocumentSummary {
  path: string;
  title: string;
  meta: Record<string, unknown>;
  mod_time: string;
  size: number;
}

export interface DocumentDetail extends DocumentSummary {
  body: string;
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface TagCount {
  tag: string;
  count: number;
}

export interface SearchResult {
  path: string;
  title: string;
  snippet: string;
  score: number;
  meta: Record<string, unknown>;
}

export interface MetadataField {
  name: string;
  type: string;
  values: string[];
}

export interface GitCommit {
  hash: string;
  author: string;
  date: string;
  message: string;
}

export interface BlameLine {
  hash: string;
  author: string;
  date: string;
  line_no: number;
  content: string;
}

export interface GraphNode {
  path: string;
  title: string;
  tags: string[];
}

export interface GraphEdge {
  source: string;
  target: string;
  type: "link" | "tag";
  label?: string;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface AppConfig {
  title: string;
  theme: string;
  themes: string[];
  font: string;
  font_url: string;
  font_family: string;
  tag_icons: TagIcon[];
}
