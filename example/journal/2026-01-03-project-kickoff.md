---
title: "プロジェクトキックオフ"
status: published
tags:
  - journal
  - milestone
author: "Taro Yamada"
priority: high
created: "2026-01-03"
updated: "2026-01-03"
---

# プロジェクトキックオフ

## 背景

チーム内のドキュメントが Notion、Confluence、ローカルの Markdown ファイルに分散しており、横断検索が困難だった。

## 決定事項

1. **Go + Preact** でシングルバイナリツールを開発
2. **読み取り専用** — ソースファイルは一切変更しない
3. **SQLite FTS5** で高速な全文検索を実現
4. **フロントマター** を検索・フィルタの基盤にする

## マイルストーン計画

| バージョン | 内容 | 目標日 |
|-----------|------|--------|
| v0.1.0 | コアエンジン（スキャナ、FTS5、REST API） | 1月末 |
| v0.2.0 | Web UI（フォルダビュー、Markdown 描画） | 2月初旬 |
| v0.3.0 | 検索 UI + ファセットフィルタ | 2月中旬 |
| v0.4.0 | Git 連携（履歴、差分、blame） | 2月中旬 |
| v0.5.0 | グラフビュー + リアルタイム通知 | 2月下旬 |

## 技術選定の理由

### なぜ Go か

- シングルバイナリでの配布が容易
- `go:embed` で SPA を埋め込める
- `modernc.org/sqlite` による Pure Go SQLite
- クロスコンパイルが容易

### なぜ Preact か

- React 互換で軽量（3KB gzip）
- 埋め込みバイナリサイズの最小化

## 関連

- [[architecture/overview]] - 技術設計の詳細
- [[guides/getting-started]] - セットアップガイド
