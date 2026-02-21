---
title: "Frontmatter ベストプラクティス"
status: published
tags:
  - tutorial
  - beginner
  - yaml
author: "Taro Yamada"
priority: medium
created: "2026-01-12"
updated: "2026-02-10"
---

# Frontmatter ベストプラクティス

YAML フロントマターを効果的に使うためのガイドラインです。

## 必須フィールド

```yaml
---
title: "ドキュメントのタイトル"
status: draft
tags:
  - カテゴリ名
---
```

## 推奨フィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `title` | string | ドキュメントのタイトル |
| `status` | string | published / draft / review / archived |
| `tags` | array | 分類用タグ（複数指定可） |
| `created` | string | 作成日 (ISO 8601) |
| `updated` | string | 最終更新日 (ISO 8601) |
| `author` | string | 著者名 |
| `priority` | string | high / medium / low |

## タグの命名規則

- **小文字のケバブケース** を使用: `web-api`, `real-time`
- **具体的に**: `go` > `programming`
- **一貫性**: チーム内で同じタグを使う
- **2〜5個** が適切な数

## ステータスの遷移

```
draft → review → published → archived
```

1. **draft**: 執筆中。不完全でも OK
2. **review**: レビュー待ち。内容は完成している
3. **published**: 公開済み。品質が保証されている
4. **archived**: 古くなった情報。検索結果では下位に

## Git 日付のフォールバック

`created` や `updated` をフロントマターに書かなくても、Git のコミット履歴から自動取得されます。明示的に日付を指定した場合はそちらが優先されます。

## 関連

- [[guides/getting-started]] - 基本的なセットアップ
- [メタデータ API](../api/rest-endpoints.md) - メタデータフィールド検出 API
