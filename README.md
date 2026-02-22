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

- **Folder View** - ディレクトリ階層をツリー表示、タグ連動の絵文字アイコン対応（`/api/v1/tree`）
- **Document Viewer** - YAML frontmatter + Markdown 本文を解析して表示
- **Full-text Search** - SQLite FTS5 trigram による BM25 ランキング付き検索
- **Tag & Metadata Filter** - frontmatter の status / tag でフィルタリング
- **Graph View** - タグ共有・内部リンクベースのドキュメント関連グラフ（`/api/v1/graph`）
- **Git Integration** - ファイル単位のコミット履歴、diff、行単位 blame
- **Live Reload** - fsnotify + WebSocket でファイル変更をブラウザへ即時反映
- **REST API** - 全機能を API で提供、AI エージェントからのプログラマティックアクセス対応
- **Read-only** - ソースファイルを一切変更しない安全設計
- **Single Binary** - Preact SPA を `go:embed` で同梱、デプロイは 1 ファイル
- **Per-Repo Theming** - リポジトリごとにテーマカラー、タイトル、フォントをカスタマイズ（VS Code 系テーマ対応）

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

## Customization

複数リポジトリで markdown-kb を同時に使うとき、テーマカラー・タイトル・フォントでインスタンスを区別できます。

### Per-Repo Config

リポジトリルートに `.markdown-kb.yml`（または `.markdown-kb.yaml`）を配置：

```yaml
title: "DevShot Docs"
theme: dracula
font: noto-sans
tag_icons:
  - tag: tech
    emoji: "💻"
  - tag: idea
    emoji: "💡"
  - tag: api
    emoji: "🔌"
```

| フィールド | 説明 | デフォルト |
|-----------|------|----------|
| `title` | ヘッダー・ブラウザタブに表示される名前 | ディレクトリ名 |
| `theme` | カラーテーマ名 | `default` |
| `font` | フォントプリセット名 | `default` |
| `tag_icons` | frontmatter タグに応じたサイドバーの絵文字アイコン | なし |

### Tag Icons

`tag_icons` を設定すると、サイドバーのファイルツリーで各ドキュメントの frontmatter タグに応じた絵文字アイコンが表示されます。

- 配列の**先頭から順にマッチ**し、最初にヒットした絵文字が使われます
- 複数タグが該当する場合は `tag_icons` 配列で先に定義されている方が優先
- マッチするタグがないファイルはアイコンなしで表示されます

### CLI フラグで上書き

```bash
kb serve --title "My Project" --theme nord --font serif
kb serve /path/to/docs --theme gruvbox --open
```

### 対応テーマ

| テーマ | 説明 |
|-------|------|
| `default` | デフォルト (Blue) |
| `tokyo-night` | Tokyo Night |
| `dracula` | Dracula (Purple) |
| `nord` | Nord (Frost Blue) |
| `solarized` | Solarized (Warm Amber) |
| `monokai` | Monokai (Green) |
| `github` | GitHub |
| `catppuccin` | Catppuccin (Lavender) |
| `gruvbox` | Gruvbox (Orange) |
| `rose-pine` | Rosé Pine (Pink) |

各テーマは Light / Dark 両モードに対応しています。ヘッダー上部にアクセントカラーのバーが表示され、一目でどのリポジトリか判別できます。

### 対応フォント

| フォント | 説明 |
|---------|------|
| `default` | BIZ UDPGothic（デフォルト） |
| `noto-sans` | Noto Sans JP（クリーンなゴシック） |
| `rounded` | M PLUS Rounded 1c（丸ゴシック） |
| `serif` | Noto Serif JP（明朝体） |
| `zen-kaku` | Zen Kaku Gothic New（モダンゴシック） |

フォントは Google Fonts から動的に読み込まれます。日本語・英語両対応。

### Config API

```bash
# 現在のテーマ設定を取得
curl localhost:3000/api/v1/config
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
