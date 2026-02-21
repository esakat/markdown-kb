import type {
  ApiResponse,
  AppConfig,
  DocumentDetail,
  PaginatedResponse,
  DocumentSummary,
  TreeNode,
  TagCount,
  SearchResult,
  MetadataField,
  GitCommit,
  BlameLine,
  GraphData,
} from "../types/api";

const BASE = "/api/v1";

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || res.statusText);
  }
  return res.json();
}

export function getTree(): Promise<ApiResponse<TreeNode>> {
  return fetchJSON(`${BASE}/tree`);
}

export function getDocument(
  path: string,
): Promise<ApiResponse<DocumentDetail>> {
  return fetchJSON(`${BASE}/documents/${encodeURI(path)}`);
}

export function listDocuments(
  page = 1,
  limit = 20,
  filters?: Record<string, string>,
): Promise<PaginatedResponse<DocumentSummary>> {
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  });
  if (filters) {
    for (const [key, val] of Object.entries(filters)) {
      if (val) params.set(key, val);
    }
  }
  return fetchJSON(`${BASE}/documents?${params}`);
}

export function searchDocuments(
  query: string,
  page = 1,
  limit = 20,
  filters?: Record<string, string>,
): Promise<PaginatedResponse<SearchResult>> {
  const params = new URLSearchParams({
    q: query,
    page: String(page),
    limit: String(limit),
  });
  if (filters) {
    for (const [key, val] of Object.entries(filters)) {
      if (val) params.set(key, val);
    }
  }
  return fetchJSON(`${BASE}/search?${params}`);
}

export function listTags(): Promise<ApiResponse<TagCount[]>> {
  return fetchJSON(`${BASE}/tags`);
}

export function getMetadataFields(): Promise<ApiResponse<MetadataField[]>> {
  return fetchJSON(`${BASE}/metadata/fields`);
}

export function getFileHistory(
  path: string,
): Promise<ApiResponse<GitCommit[]>> {
  return fetchJSON(`${BASE}/git/history/${encodeURI(path)}`);
}

export function getFileDiff(
  path: string,
  from: string,
  to: string,
): Promise<ApiResponse<string>> {
  return fetchJSON(
    `${BASE}/git/diff/${encodeURI(path)}?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
  );
}

export function getFileBlame(
  path: string,
  start?: number,
  end?: number,
): Promise<ApiResponse<BlameLine[]>> {
  const params = new URLSearchParams();
  if (start !== undefined && end !== undefined) {
    params.set("start", String(start));
    params.set("end", String(end));
  }
  const qs = params.toString();
  return fetchJSON(`${BASE}/git/blame/${encodeURI(path)}${qs ? `?${qs}` : ""}`);
}

export function getGraph(): Promise<ApiResponse<GraphData>> {
  return fetchJSON(`${BASE}/graph`);
}

export function getConfig(): Promise<AppConfig> {
  return fetchJSON(`${BASE}/config`);
}
