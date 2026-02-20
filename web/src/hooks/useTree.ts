import { useApi } from "./useApi";
import { getTree } from "../api/client";
import type { TreeNode } from "../types/api";

export function useTree() {
  const { data, loading, error } = useApi(() => getTree());

  return {
    tree: data?.data ?? null,
    loading,
    error,
  };
}

export type { TreeNode };
