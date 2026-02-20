import type {
  ApiResponse,
  DocumentDetail,
  PaginatedResponse,
  DocumentSummary,
  TreeNode,
  TagCount,
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

export function getDocument(path: string): Promise<ApiResponse<DocumentDetail>> {
  return fetchJSON(`${BASE}/documents/${encodeURI(path)}`);
}

export function listDocuments(
  page = 1,
  limit = 20
): Promise<PaginatedResponse<DocumentSummary>> {
  return fetchJSON(`${BASE}/documents?page=${page}&limit=${limit}`);
}

export function searchDocuments(
  query: string,
  page = 1,
  limit = 20
): Promise<PaginatedResponse<DocumentSummary>> {
  return fetchJSON(
    `${BASE}/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`
  );
}

export function listTags(): Promise<ApiResponse<TagCount[]>> {
  return fetchJSON(`${BASE}/tags`);
}
