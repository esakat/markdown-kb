---
title: "AI エージェントでの活用ガイド"
status: draft
tags:
  - ai
  - api
  - tutorial
author: "Taro Yamada"
priority: medium
created: "2026-02-18"
updated: "2026-02-20"
---

# AI エージェントでの活用ガイド

Markdown KB を AI エージェント（Claude Code、Cursor、Copilot 等）のコンテキストソースとして活用する方法。

## なぜ KB を使うのか

AI エージェントは大量のコンテキストを扱えますが、**関連性の高い情報を効率的に取得** することが重要です。Markdown KB は:

- 全文検索で関連ドキュメントを素早く発見
- メタデータフィルタで精度の高い絞り込み
- グラフ構造で関連性を辿る

## セットアップ

```bash
# 1. KB をバックグラウンドで起動
kb serve ./docs --port 3000 &

# 2. 動作確認
curl -s localhost:3000/api/v1/health
# → {"status":"ok"}
```

## 検索パターン

### キーワード検索

```bash
curl -s 'localhost:3000/api/v1/search?q=認証+設計&limit=3' | jq '.data[]'
```

### タグ + ステータスでフィルタ

```bash
curl -s 'localhost:3000/api/v1/documents?tag=architecture&status=published' | jq '.data[].path'
```

### 関連ドキュメントの探索

```bash
# 1. まず検索で起点を見つける
DOC=$(curl -s 'localhost:3000/api/v1/search?q=WebSocket&limit=1' | jq -r '.data[0].path')

# 2. グラフでリンク先を取得
curl -s 'localhost:3000/api/v1/graph' | \
  jq --arg d "$DOC" '[.data.edges[] | select(.source == $d) | .target]'
```

## MCP サーバとして（将来構想）

将来的には MCP (Model Context Protocol) サーバとして、AI エージェントが直接 KB にアクセスできるようにする構想があります。

```
Claude Code ←→ MCP Protocol ←→ Markdown KB
```

これにより `curl` を介さず、ネイティブなツール呼び出しで KB を検索できるようになります。

## 関連

- [[api/rest-endpoints]] - API 仕様
- [[recipes/claude-code-integration]] - Claude Code 連携レシピ
- [[guides/search-tips]] - 検索のコツ
