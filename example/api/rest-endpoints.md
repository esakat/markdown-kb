---
title: "REST API リファレンス"
status: published
tags:
  - api
  - go
  - reference
author: "Taro Yamada"
priority: high
created: "2026-01-08"
updated: "2026-02-20"
---

# REST API リファレンス

Markdown KB は読み取り専用の REST API を提供します。AI エージェント（Claude Code 等）からのプログラマティックアクセスに最適化されています。

## ベース URL

```
http://localhost:3000/api/v1
```

## エンドポイント一覧

### ドキュメント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/documents` | ドキュメント一覧（ページネーション付き） |
| GET | `/documents/{path}` | ドキュメント詳細（本文含む） |
| GET | `/search?q=keyword` | 全文検索 |
| GET | `/tree` | ディレクトリツリー |
| GET | `/tags` | タグ一覧（カウント付き） |
| GET | `/metadata/fields` | メタデータフィールド自動検出 |

### Git 連携

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/git/history/{path}` | ファイルのコミット履歴 |
| GET | `/git/diff/{path}?from=a&to=b` | 2コミット間の差分 |
| GET | `/git/blame/{path}` | 行ごとの最終変更者 |

### グラフ

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/graph` | ドキュメントグラフ（ノード + エッジ） |

### その他

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/raw/{path}` | 生ファイル配信 |
| GET | `/ws` | WebSocket（リアルタイム通知） |
| GET | `/health` | ヘルスチェック |

## レスポンス形式

### 単一オブジェクト

```json
{
  "data": { ... }
}
```

### ページネーション

```json
{
  "data": [ ... ],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

### エラー

```json
{
  "error": "document not found"
}
```

## 使用例

### Claude Code から検索

```bash
curl -s 'localhost:3000/api/v1/search?q=認証&limit=5' | jq '.data[].title'
```

### フィルタ付きリスト

```bash
curl -s 'localhost:3000/api/v1/documents?status=published&tag=go' | jq '.data[].path'
```

### グラフデータ取得

```bash
curl -s 'localhost:3000/api/v1/graph' | jq '.data.nodes | length'
```

## 関連

- [[guides/getting-started]] - セットアップ
- [[guides/search-tips]] - 検索のコツ
- [[architecture/overview]] - アーキテクチャ概要
