---
title: "FTS5 検索エンジン設計"
status: published
tags:
  - sqlite
  - search
  - architecture
author: "Taro Yamada"
priority: high
created: "2026-01-05"
updated: "2026-02-18"
---

# FTS5 検索エンジン設計

SQLite FTS5 を利用した全文検索の設計を解説します。

## テーブル構造

### メインテーブル (`documents`)

```sql
CREATE TABLE documents (
  path  TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body  TEXT NOT NULL,
  meta  TEXT NOT NULL,  -- JSON
  mtime TEXT NOT NULL
);
```

### FTS5 仮想テーブル (`documents_fts`)

```sql
CREATE VIRTUAL TABLE documents_fts USING fts5(
  title,
  body,
  content='documents',
  content_rowid='rowid',
  tokenize='trigram'
);
```

## トークナイザ選定

| トークナイザ | 日本語対応 | 速度 | 精度 |
|-------------|-----------|------|------|
| unicode61   | △ 単語境界不正確 | ◎ | △ |
| trigram     | ◎ N-gram ベース | ○ | ◎ |
| porter      | × 英語専用 | ◎ | ○ |

**trigram を採用**: 日本語・英語混在テキストで安定した検索結果を提供。

## BM25 スコアリング

FTS5 組み込みの `bm25()` 関数を使用:

```sql
SELECT path, title,
       snippet(documents_fts, 1, '<b>', '</b>', '...', 32) AS snippet,
       bm25(documents_fts, 10.0, 1.0) AS score
FROM documents_fts
WHERE documents_fts MATCH ?
ORDER BY score
LIMIT ? OFFSET ?;
```

- タイトル一致に **10倍** の重みを設定
- スコアは負の値（0 に近いほど高関連度）
- `snippet()` でマッチ箇所を `<b>` タグでハイライト

## メタデータフィルタリング

FTS5 と JSON 関数を組み合わせたハイブリッド検索:

```sql
SELECT d.path, d.title, ...
FROM documents_fts f
JOIN documents d ON d.rowid = f.rowid
WHERE documents_fts MATCH ?
  AND json_extract(d.meta, '$.status') = ?
  AND EXISTS (
    SELECT 1 FROM json_each(json_extract(d.meta, '$.tags'))
    WHERE value = ?
  )
ORDER BY bm25(documents_fts, 10.0, 1.0);
```

## インデックス更新戦略

### UPSERT パターン

```sql
INSERT INTO documents (path, title, body, meta, mtime)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
  title = excluded.title,
  body  = excluded.body,
  meta  = excluded.meta,
  mtime = excluded.mtime;
```

### FTS5 同期トリガー

```sql
CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
  INSERT INTO documents_fts(rowid, title, body)
  VALUES (new.rowid, new.title, new.body);
END;
```

## パフォーマンス

| ドキュメント数 | インデックス構築 | 検索レスポンス |
|--------------|----------------|---------------|
| 100          | < 50ms         | < 5ms         |
| 1,000        | < 500ms        | < 10ms        |
| 10,000       | < 5s           | < 20ms        |

## 関連

- [[architecture/overview]] - システム全体の設計
- [[guides/search-tips]] - ユーザ向け検索ガイド
- [[api/rest-endpoints]] - 検索 API 仕様
