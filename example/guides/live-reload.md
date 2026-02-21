---
title: "ライブリロード機能"
status: published
tags:
  - websocket
  - real-time
  - tutorial
author: "Taro Yamada"
priority: low
created: "2026-02-10"
updated: "2026-02-20"
---

# ライブリロード機能

ファイルを保存すると、ブラウザが自動的に更新されます。

## 仕組み

```
エディタで保存
  ↓
fsnotify がファイル変更を検知
  ↓
インデックスを差分更新（追加/変更/削除）
  ↓
WebSocket で全クライアントに通知
  ↓
ブラウザにトースト通知 + データ再取得
```

## 通知の種類

| イベント | トースト色 | 説明 |
|---------|-----------|------|
| created | 緑 | 新しい .md ファイルが追加された |
| updated | 青 | 既存の .md ファイルが更新された |
| deleted | オレンジ | .md ファイルが削除された |

## 対象ファイル

- `.md` 拡張子のファイルのみ監視
- `.git/`, `node_modules/`, ドットディレクトリは除外
- サブディレクトリの追加も自動検知

## デバウンス

同じファイルへの連続した変更は **300ms** のデバウンスで集約されます。エディタの自動保存など、短時間に複数回の書き込みが発生しても通知は 1 回です。

## WebSocket エンドポイント

```javascript
const ws = new WebSocket('ws://localhost:3000/api/v1/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data.type, data.path);
  // → "updated" "guides/getting-started.md"
};
```

## 関連

- [[architecture/websocket-design]] - WebSocket の設計詳細
- [[guides/getting-started]] - 基本的なセットアップ
