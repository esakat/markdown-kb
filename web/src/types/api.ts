export interface TreeNode {
  name: string;
  type: "dir" | "file";
  path?: string;
  title?: string;
  children?: TreeNode[];
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
