# Markdown KB

Git リポジトリ内の YAML frontmatter 付き Markdown ファイルを、ブラウザで閲覧・検索できるシングルバイナリ Web アプリ。

人間がブラウズしやすく、AI エージェント（Claude Code 等）が REST API でプログラマティックに検索・参照できる「人と AI 双方にフレンドリー」なナレッジベースビューア。

## Quick Start

```bash
# Build
make build

# Serve current directory
./kb serve

# Serve specific directory on custom port
./kb serve /path/to/docs --port 8080 --open

# Build search index (CI usage)
./kb index --format json
```

## Features

- **Folder View** - Git リポジトリのディレクトリ階層をツリー表示
- **Graph/Tag View** - Scrapbox ライクなタグベースのドキュメント関連付け
- **Full-text Search** - SQLite FTS5 (trigram) による BM25 ランキング付き検索
- **Git Integration** - コミット履歴、diff 表示、オンデマンド blame
- **REST API** - AI エージェント向けプログラマティックアクセス
- **Read-only** - ソースファイルを一切変更しない安全設計

## API

```bash
# List documents
curl localhost:3000/api/v1/documents

# Filter by metadata
curl localhost:3000/api/v1/documents?status=spec&tag=ai

# Full-text search
curl localhost:3000/api/v1/search?q=習慣トラッカー

# Discover metadata schema
curl localhost:3000/api/v1/metadata/fields
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go (single binary, `//go:embed`) |
| Web UI | Preact + D3.js |
| Frontmatter | goccy/go-yaml |
| Markdown | goldmark |
| Search | SQLite FTS5 via modernc.org/sqlite |
| Git | go-git + git CLI |
| File Watch | fsnotify |

## Development

```bash
# Run tests
make test

# Run with coverage
make cover

# Lint
make lint
```

## License

MIT
