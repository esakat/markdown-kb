---
title: "Getting Started with Markdown KB"
status: published
tags:
  - go
  - tutorial
  - beginner
author: "Taro Yamada"
priority: high
created: "2026-01-05"
updated: "2026-02-15"
---

# Getting Started with Markdown KB

Markdown KB は Git リポジトリ内の Markdown ファイルをブラウズ・検索するためのツールです。

## 前提条件

- Go 1.21 以上
- Node.js 20 以上（開発時のみ）
- Git

## インストール

### バイナリから

```bash
# GitHub Releases からダウンロード
curl -LO https://github.com/esakat/markdown-kb/releases/latest/download/kb_linux_amd64.tar.gz
tar xzf kb_linux_amd64.tar.gz
sudo mv kb /usr/local/bin/
```

### ソースから

```bash
git clone https://github.com/esakat/markdown-kb.git
cd markdown-kb
make build
```

## 基本的な使い方

```bash
# カレントディレクトリを配信
kb serve

# 指定ディレクトリをポート 8080 で配信
kb serve /path/to/docs --port 8080 --open

# インデックス情報を JSON で出力
kb index /path/to/docs --format json
```

## Frontmatter の書き方

YAML フロントマターで文書のメタデータを定義できます:

```yaml
---
title: "ドキュメントのタイトル"
status: published
tags:
  - go
  - tutorial
created: "2026-01-01"
updated: "2026-02-20"
---
```

## 次のステップ

- API の使い方は [[api/rest-endpoints]] を参照
- アーキテクチャの全体像は [設計概要](../architecture/overview.md) を参照
- 検索の仕組みは [[guides/search-tips]] で詳しく解説
