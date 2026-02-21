---
title: "検索のコツと高度な使い方"
status: published
tags:
  - sqlite
  - search
  - tutorial
author: "Hanako Suzuki"
priority: medium
created: "2026-01-10"
updated: "2026-02-18"
---

# 検索のコツと高度な使い方

Markdown KB は SQLite FTS5 (trigram tokenizer) を使った全文検索を提供します。

## 基本検索

検索バーにキーワードを入力するだけで、タイトル・本文・メタデータを横断検索できます。

- **日本語対応**: trigram トークナイザにより日本語の検索も可能
- **BM25 ランキング**: 関連度の高い結果が上位に表示される
- **スニペット**: 検索ヒット箇所がハイライト表示される

## フィルタリング

### ステータスフィルタ

ドキュメントの `status` フィールドでフィルタリングできます:

| ステータス | 用途 |
|-----------|------|
| `published` | 公開済みの完成ドキュメント |
| `draft` | 執筆中の下書き |
| `review` | レビュー待ち |
| `archived` | アーカイブ済み |

### タグフィルタ

左パネルのタグバッジをクリックして、特定のタグを持つドキュメントに絞り込めます。

## 検索 URL の共有

検索状態は URL に反映されます:

```
/search?q=sqlite&status=published&tag=tutorial&page=1
```

この URL をチームメンバーに共有すれば、同じ検索結果を再現できます。

## API からの検索

```bash
# 基本検索
curl 'localhost:3000/api/v1/search?q=検索キーワード'

# フィルタ付き検索
curl 'localhost:3000/api/v1/search?q=Go&status=published&tag=tutorial'
```

レスポンスには BM25 スコアとスニペットが含まれます。詳細は [[api/rest-endpoints]] を参照。

## 関連

- [[guides/getting-started]] - 基本的なセットアップ
- [[architecture/fts5-design]] - FTS5 の内部設計
