---
title: "API レスポンスフォーマット詳細"
status: review
tags:
  - api
  - reference
author: "Hanako Suzuki"
priority: medium
created: "2026-01-20"
updated: "2026-02-12"
---

# API レスポンスフォーマット詳細

各エンドポイントのレスポンス構造を詳しく解説します。

## DocumentSummary

`GET /api/v1/documents` のレスポンス要素:

```json
{
  "path": "guides/getting-started.md",
  "title": "Getting Started",
  "meta": {
    "status": "published",
    "tags": ["go", "tutorial"],
    "author": "Taro Yamada"
  },
  "mod_time": "2026-02-15T10:30:00+09:00",
  "size": 2048
}
```

## DocumentDetail

`GET /api/v1/documents/{path}` のレスポンス:

```json
{
  "data": {
    "path": "guides/getting-started.md",
    "title": "Getting Started",
    "meta": { "..." },
    "body": "# Getting Started\n\n...",
    "mod_time": "2026-02-15T10:30:00+09:00",
    "size": 2048
  },
  "git_dates": {
    "created": "2026-01-05T09:00:00+09:00",
    "updated": "2026-02-15T10:30:00+09:00"
  }
}
```

`git_dates` はフロントマターに `created` / `updated` がない場合のみ付与されます。

## SearchResult

`GET /api/v1/search?q=...` のレスポンス要素:

```json
{
  "path": "guides/search-tips.md",
  "title": "検索のコツと高度な使い方",
  "snippet": "...SQLite <b>FTS5</b> を使った全文<b>検索</b>を提供...",
  "score": -3.14,
  "meta": { "status": "published", "tags": ["sqlite", "search"] }
}
```

- `snippet` の `<b>` タグでマッチ箇所がハイライトされる
- `score` は BM25 スコア（負の値。0 に近いほど関連度が高い）

## GraphData

`GET /api/v1/graph` のレスポンス:

```json
{
  "data": {
    "nodes": [
      { "path": "guides/getting-started.md", "title": "Getting Started", "tags": ["go", "tutorial"] }
    ],
    "edges": [
      { "source": "guides/getting-started.md", "target": "api/rest-endpoints.md", "type": "link" },
      { "source": "guides/getting-started.md", "target": "guides/search-tips.md", "type": "tag", "label": "tutorial" }
    ]
  }
}
```

## 関連

- [REST API リファレンス](rest-endpoints.md) - エンドポイント一覧
- [[architecture/overview]] - 技術設計
