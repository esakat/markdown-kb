---
title: "Markdown KB サンプルドキュメント"
status: published
tags:
  - index
  - overview
created: "2026-01-01"
updated: "2026-02-20"
---

# Markdown KB サンプルドキュメント

このディレクトリは **Markdown KB** の全機能を確認するためのサンプルデータです。

## ディレクトリ構成

| フォルダ | 内容 |
|---------|------|
| `guides/` | 技術ガイド・チュートリアル |
| `api/` | API 仕様書 |
| `architecture/` | アーキテクチャ設計ドキュメント |
| `journal/` | 開発日誌 |
| `recipes/` | 逆引きレシピ集 |

## 機能カバレッジ

- **フロントマター**: title, status, tags, created, updated, priority, author
- **検索**: 日本語・英語の混在コンテンツ
- **フィルタリング**: published / draft / archived / review の各ステータス
- **リンク**: `[[wiki-link]]` と `[text](path.md)` の両方
- **タグ**: go, preact, sqlite, d3, websocket, testing, security, performance 等
- **Markdown**: 表、コードブロック、リスト、見出し、数式風テキスト

## クイックスタート

```bash
# このディレクトリを KB で配信
./kb serve ./example --open
```

関連ドキュメント: [[guides/getting-started]] | [[architecture/overview]]
