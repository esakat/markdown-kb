---
title: "Docker デプロイメントレシピ"
status: archived
tags:
  - recipe
  - docker
  - deployment
author: "Hanako Suzuki"
priority: low
created: "2026-01-20"
updated: "2026-02-01"
---

# Docker デプロイメントレシピ

> **Note**: シングルバイナリ配布に移行したため、このレシピはアーカイブ済みです。goreleaser による配布を推奨します。

## Dockerfile

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /kb ./cmd/kb

FROM alpine:3.19
RUN apk add --no-cache git
COPY --from=builder /kb /usr/local/bin/kb
ENTRYPOINT ["kb"]
```

## docker-compose.yml

```yaml
version: "3.8"
services:
  kb:
    build: .
    ports:
      - "3000:3000"
    volumes:
      - ./docs:/data:ro
    command: ["serve", "/data", "--port", "3000"]
```

## 実行

```bash
docker compose up -d
open http://localhost:3000
```

## 代替: goreleaser バイナリの利用

現在はシングルバイナリでの配布を推奨しています:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/user/markdown-kb/releases/latest/download/kb_darwin_arm64 -o kb
chmod +x kb
./kb serve ./docs --open
```

## 関連

- [[guides/getting-started]] - 基本セットアップ
- [[architecture/overview]] - アーキテクチャ概要
