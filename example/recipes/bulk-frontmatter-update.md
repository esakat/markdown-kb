---
title: "フロントマター一括更新レシピ"
status: published
tags:
  - recipe
  - yaml
  - automation
author: "Hanako Suzuki"
priority: medium
created: "2026-02-08"
updated: "2026-02-12"
---

# フロントマター一括更新レシピ

既存の Markdown ファイルにフロントマターを一括追加・更新する方法。

## ユースケース

- 全ファイルに `status: draft` を追加したい
- `author` フィールドを一括で設定したい
- タグの名前を変更したい（例: `golang` → `go`）

## 方法 1: sed でシンプルに追加

```bash
# status フィールドがないファイルに追加
for f in docs/**/*.md; do
  if ! grep -q '^status:' "$f"; then
    sed -i '' '/^---$/a\
status: draft' "$f"
  fi
done
```

## 方法 2: yq で安全に操作

```bash
# yq を使って YAML フロントマターを安全に編集
for f in docs/**/*.md; do
  # フロントマター部分を抽出
  frontmatter=$(sed -n '/^---$/,/^---$/p' "$f" | sed '1d;$d')

  # yq で更新
  updated=$(echo "$frontmatter" | yq '.author = "Team"')

  # ファイルに書き戻し（本文は維持）
  body=$(sed '1,/^---$/!d; /^---$/,$!d' "$f" | tail -n +2)
  echo -e "---\n${updated}\n---\n${body}" > "$f"
done
```

## 方法 3: Markdown KB API で確認してから更新

```bash
# 1. status が未設定のファイルを KB API で検索
curl -s 'localhost:3000/api/v1/documents?limit=100' | \
  jq -r '.data[] | select(.meta.status == null) | .path'

# 2. 確認後、対象ファイルにのみ追加
# （API は読み取り専用なので、更新は直接ファイル操作）
```

## 注意事項

- **バックアップを取ってから実行** すること
- Markdown KB は読み取り専用なので、変更は外部ツールで行う
- ファイルを変更すると fsnotify が検知し、インデックスが自動更新される

## 関連

- [[guides/writing-frontmatter]] - フロントマターの書き方
- [[guides/getting-started]] - セットアップ
