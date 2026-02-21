# Markdown KB

Git リポジトリ内の YAML frontmatter 付き Markdown ファイルを、ブラウザで閲覧・検索できるシングルバイナリ Web アプリ。

人間がブラウズしやすく、AI エージェント（Claude Code 等）が REST API でプログラマティックに検索・参照できる「人と AI 双方にフレンドリー」なナレッジベースビューア。

## Install

```bash
go install github.com/esakat/markdown-kb/cmd/kb@latest
```

または [GitHub Releases](https://github.com/esakat/markdown-kb/releases) からバイナリをダウンロード。

## Quick Start

```bash
# カレントディレクトリを配信
kb serve

# ディレクトリ指定 + ポート変更 + ブラウザ自動オープン
kb serve /path/to/docs --port 8080 --open

# 検索インデックスをビルドして出力（CI 連携向け）
kb index --format json
kb index --format text
```

## Features

- **Folder View** - ディレクトリ階層をツリー表示（`/api/v1/tree`）
- **Document Viewer** - YAML frontmatter + Markdown 本文を解析して表示
- **Full-text Search** - SQLite FTS5 trigram による BM25 ランキング付き検索
- **Tag & Metadata Filter** - frontmatter の status / tag でフィルタリング
- **Graph View** - タグ共有・内部リンクベースのドキュメント関連グラフ（`/api/v1/graph`）
- **Git Integration** - ファイル単位のコミット履歴、diff、行単位 blame
- **Live Reload** - fsnotify + WebSocket でファイル変更をブラウザへ即時反映
- **REST API** - 全機能を API で提供、AI エージェントからのプログラマティックアクセス対応
- **Read-only** - ソースファイルを一切変更しない安全設計
- **Single Binary** - Preact SPA を `go:embed` で同梱、デプロイは 1 ファイル

## API

すべてのエンドポイントは JSON レスポンスを返します。ページネーション対応（`?page=1&limit=20`）。

### Documents

```bash
# ドキュメント一覧（ページネーション付き）
curl localhost:3000/api/v1/documents?page=1&limit=20

# metadata でフィルタ
curl localhost:3000/api/v1/documents?status=spec&tag=ai

# ドキュメント詳細（本文 + Git 日付補完）
curl localhost:3000/api/v1/documents/path/to/file.md

# 生ファイル取得
curl localhost:3000/api/v1/raw/path/to/file.md
```

### Search

```bash
# 全文検索（BM25 スコア + スニペット付き）
curl localhost:3000/api/v1/search?q=キーワード

# 検索 + メタデータフィルタ
curl localhost:3000/api/v1/search?q=キーワード&status=spec&tag=ai
```

### Tags & Metadata

```bash
# タグ一覧（出現回数付き）
curl localhost:3000/api/v1/tags

# frontmatter フィールドのスキーマ自動検出
curl localhost:3000/api/v1/metadata/fields
```

### Structure

```bash
# ディレクトリツリー
curl localhost:3000/api/v1/tree

# ドキュメント関連グラフ（タグ共有 + 内部リンク）
curl localhost:3000/api/v1/graph
```

### Git

```bash
# ファイルのコミット履歴
curl localhost:3000/api/v1/git/history/path/to/file.md

# コミット間 diff
curl localhost:3000/api/v1/git/diff/path/to/file.md?from=abc123&to=def456

# 行単位 blame（範囲指定可）
curl localhost:3000/api/v1/git/blame/path/to/file.md
curl localhost:3000/api/v1/git/blame/path/to/file.md?start=10&end=20
```

### Other

```bash
# ヘルスチェック
curl localhost:3000/api/health

# WebSocket（ライブリロード用）
wscat -c ws://localhost:3000/api/v1/ws
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go (single binary, `//go:embed`) |
| CLI | Cobra |
| Web UI | Preact + D3.js |
| Frontmatter | goccy/go-yaml |
| Markdown | goldmark |
| Search | SQLite FTS5 trigram (modernc.org/sqlite, Pure Go) |
| Git | git CLI (history, diff, blame) |
| File Watch | fsnotify (debounced, recursive) |
| WebSocket | nhooyr.io/websocket |

## Development

```bash
# ソースからビルド
make build

# テスト実行
make test

# カバレッジ付きテスト
make cover

# Lint
make lint
```

## License

MIT
