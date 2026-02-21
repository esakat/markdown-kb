---
title: "WebSocket リアルタイム通知設計"
status: published
tags:
  - websocket
  - architecture
  - real-time
author: "Taro Yamada"
priority: medium
created: "2026-01-10"
updated: "2026-02-15"
---

# WebSocket リアルタイム通知設計

ファイル変更をブラウザにリアルタイム通知する仕組みを解説します。

## 設計概要

```
fsnotify (OS イベント)
    ↓
Debouncer (300ms)
    ↓
Re-index (SQLite UPSERT)
    ↓
Hub.Broadcast(WSEvent)
    ↓ (fan-out)
Client 1, Client 2, ... Client N
```

## Hub パターン

### クライアント管理

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[*websocket.Conn]context.CancelFunc
}
```

- `sync.RWMutex` でスレッドセーフなクライアント管理
- 各クライアントに `context.CancelFunc` を紐付け
- 切断時は自動的にマップから除外

### イベント構造

```json
{
  "type": "updated",
  "path": "guides/getting-started.md"
}
```

| type | 説明 |
|------|------|
| `created` | 新規 `.md` ファイル追加 |
| `updated` | 既存ファイル更新 |
| `deleted` | ファイル削除 |

## 接続フロー

1. クライアントが `ws://host/api/v1/ws` に接続
2. `nhooyr.io/websocket` で WebSocket にアップグレード
3. Hub がクライアントを登録
4. サーバがブロッキング read ループで接続維持
5. ファイル変更イベント → Hub が全クライアントに JSON 送信
6. クライアントがトースト通知を表示 + データ再取得

## 再接続戦略

フロントエンドでは自動再接続を実装:

```typescript
const RECONNECT_DELAY = 3000; // 3秒

function connect() {
  const ws = new WebSocket(url);
  ws.onclose = () => {
    setTimeout(connect, RECONNECT_DELAY);
  };
}
```

## デバウンス設計

```go
var timers sync.Map // map[string]*time.Timer

func debounce(path string, fn func()) {
    if t, ok := timers.Load(path); ok {
        t.(*time.Timer).Stop()
    }
    timer := time.AfterFunc(300*time.Millisecond, func() {
        timers.Delete(path)
        fn()
    })
    timers.Store(path, timer)
}
```

- ファイルパスごとに独立したタイマー
- 300ms 以内の連続変更は最後の 1 回に集約
- エディタの自動保存に対応

## 関連

- [[guides/live-reload]] - ユーザ向けライブリロードガイド
- [[architecture/overview]] - システム全体像
- [[api/rest-endpoints]] - WebSocket エンドポイント仕様
