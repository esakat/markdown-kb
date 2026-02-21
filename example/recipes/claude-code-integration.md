---
title: "Claude Code との連携レシピ"
status: published
tags:
  - recipe
  - ai
  - api
  - automation
author: "Taro Yamada"
priority: high
created: "2026-02-10"
updated: "2026-02-20"
---

# Claude Code との連携レシピ

Markdown KB の REST API を Claude Code から活用する方法を紹介します。

## 基本パターン

### ドキュメント検索

```bash
# KB で「認証」に関するドキュメントを検索
curl -s 'localhost:3000/api/v1/search?q=認証' | jq '.data[].title'
```

### タグベースの探索

```bash
# "architecture" タグのドキュメント一覧
curl -s 'localhost:3000/api/v1/documents?tag=architecture' | jq '.data[].path'
```

### ドキュメント内容の取得

```bash
# 特定ドキュメントの本文を取得
curl -s 'localhost:3000/api/v1/documents/guides/getting-started.md' | jq -r '.data.body'
```

## 高度なパターン

### グラフから関連ドキュメントを探す

```bash
# あるドキュメントにリンクしている全ドキュメントを取得
TARGET="architecture/overview.md"
curl -s 'localhost:3000/api/v1/graph' | \
  jq --arg t "$TARGET" '[.data.edges[] | select(.target == $t) | .source]'
```

### メタデータフィールドの自動検出

```bash
# どんなフロントマターフィールドが使われているか確認
curl -s 'localhost:3000/api/v1/metadata/fields' | jq '.data'
```

### ディレクトリ構造の把握

```bash
# ツリー構造を取得
curl -s 'localhost:3000/api/v1/tree' | jq '.'
```

## CLAUDE.md での活用

プロジェクトの `CLAUDE.md` に KB サーバの情報を記載すると、Claude Code が自動的にナレッジベースを参照できます:

```markdown
## Knowledge Base
- KB Server: http://localhost:3000
- Search: curl -s 'localhost:3000/api/v1/search?q=KEYWORD' | jq '.data'
- Docs: curl -s 'localhost:3000/api/v1/documents' | jq '.data[].path'
```

## 関連

- [[api/rest-endpoints]] - REST API 仕様
- [[api/response-format]] - レスポンスフォーマット
- [[guides/search-tips]] - 検索のコツ
