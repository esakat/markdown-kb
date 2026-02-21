---
title: "アーキテクチャ概要"
status: published
tags:
  - go
  - architecture
  - sqlite
author: "Taro Yamada"
priority: high
created: "2026-01-03"
updated: "2026-02-20"
---

# アーキテクチャ概要

Markdown KB のシステム全体像を説明します。

## 設計原則

1. **読み取り専用** — ソースファイルを一切変更しない
2. **シングルバイナリ** — SPA を `go:embed` で埋め込み
3. **フロントマターファースト** — YAML メタデータが検索・フィルタの基盤
4. **ハイブリッド Git** — go-git で構造解析、git CLI で blame/diff

## コンポーネント図

```
┌─────────────┐     ┌─────────────┐
│  Web Browser │◄────│  Preact SPA │
│  (User/AI)   │     │  (embedded) │
└──────┬───────┘     └──────┬──────┘
       │ HTTP/WS            │
┌──────▼───────────────────▼──────┐
│           Go HTTP Server         │
│  ┌──────────┐  ┌──────────────┐  │
│  │ REST API │  │  WebSocket   │  │
│  │ handlers │  │    Hub       │  │
│  └────┬─────┘  └──────┬──────┘  │
│       │                │         │
│  ┌────▼────────────────▼─────┐  │
│  │     SQLite FTS5 Store     │  │
│  └────────────┬──────────────┘  │
│               │                  │
│  ┌────────────▼──────────────┐  │
│  │    Scanner + Parser       │  │
│  └────────────┬──────────────┘  │
│               │                  │
│  ┌────────────▼──────────────┐  │
│  │    fsnotify Watcher       │  │
│  └───────────────────────────┘  │
└──────────────────────────────────┘
       │
┌──────▼───────┐
│  File System  │
│  (.md files)  │
└──────────────┘
```

## パッケージ構成

| パッケージ | 責務 |
|-----------|------|
| `cmd/kb` | CLI エントリポイント (Cobra) |
| `internal/scanner` | ディレクトリ走査、.md ファイル収集 |
| `internal/parser` | YAML フロントマター解析、リンク抽出 |
| `internal/index` | SQLite FTS5 インデックス、グラフ構築 |
| `internal/server` | HTTP ルーティング、WebSocket Hub |
| `internal/git` | Git CLI ラッパー（history/diff/blame） |
| `internal/watcher` | fsnotify ファイル監視 |
| `web/` | Preact SPA (Vite + TypeScript) |

## データフロー

### 起動時

1. `scanner.Scan()` で全 `.md` ファイルを収集
2. `parser.ParseFrontmatter()` で YAML 解析
3. `store.IndexDocument()` で SQLite に UPSERT
4. `watcher.Start()` で監視開始
5. `server.Start()` で HTTP サーバ起動

### ファイル変更時

1. fsnotify がイベント検知（300ms デバウンス）
2. 変更されたファイルを再パース
3. インデックスを差分更新
4. WebSocket Hub 経由で全クライアントに通知

## 関連

- [[architecture/fts5-design]] - 検索エンジンの詳細設計
- [[architecture/websocket-design]] - WebSocket の設計
- [[api/rest-endpoints]] - API 仕様
- [[guides/getting-started]] - 使い始め方
