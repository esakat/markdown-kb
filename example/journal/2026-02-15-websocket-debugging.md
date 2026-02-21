---
title: "WebSocket デバッグ記録"
status: draft
tags:
  - journal
  - websocket
  - debugging
author: "Taro Yamada"
priority: low
created: "2026-02-15"
updated: "2026-02-15"
---

# WebSocket デバッグ記録

## 問題

テスト環境で WebSocket クライアントが 0 件と報告される。`Hub.ClientCount()` が常に 0 を返す。

## 調査

### 仮説 1: 接続自体が失敗している

→ ログを確認したが、接続は成功していた。

### 仮説 2: ServeWS がすぐに return している

→ **正解**。`ServeWS` 内で `go func()` でリードループを起動した直後に return していた。

```go
// 問題のあったコード
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
    conn, _ := websocket.Accept(w, r, nil)
    h.addClient(conn)
    go h.readLoop(conn) // goroutine を起動して即 return
    // ← ここで HTTP ハンドラが終了し、テストサーバが接続を閉じる
}
```

### 修正

```go
// 修正後
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
    conn, _ := websocket.Accept(w, r, nil)
    h.addClient(conn)
    h.readLoop(conn) // ブロッキングで実行
}
```

## 学び

- `httptest.Server` はハンドラが return すると接続をクリーンアップする
- WebSocket ハンドラはリードループでブロックする必要がある
- 本番環境では `http.Server` が goroutine を生成するので問題ない

## 関連

- [[architecture/websocket-design]] - WebSocket 設計
- [[guides/live-reload]] - ライブリロード機能
