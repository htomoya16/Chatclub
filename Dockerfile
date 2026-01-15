# syntax=docker/dockerfile:1.10

####################
# ビルドステージ（prod 共通）
####################
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0

RUN apk add --no-cache \
            git \
            ca-certificates \
            tzdata

# Go モジュールファイルをコピー
COPY go.mod go.sum ./
# 依存レイヤのキャッシュ
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# ソースコードをコピー
COPY . .

# ビルド（静的リンク）　絶対パスをバイナリから取り除く
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -o /app/main ./cmd/server

####################
# 実行ステージ（prod）
####################
FROM alpine:3.20 AS prod

WORKDIR /app

# パッケージのインストール（--no-cacheでキャッシュ削除）
RUN apk add --no-cache \
            ca-certificates \
            tzdata

# バイナリをコピー
COPY --from=builder /app/main ./main

# 非rootユーザ
RUN adduser -D -u 65532 user
USER user

# ポートを動的に設定
ENV PORT=8080
EXPOSE 8080

CMD ["./main"]

####################
# 開発ステージ（dev ホットリロード）
####################
FROM golang:1.25.1-alpine AS dev

WORKDIR /app

# dev はビルドやデバッグに必要なツールを入れる
RUN apk add --no-cache \
            git \
            ca-certificates \
            tzdata \
            build-base \
            bash \
            && go install github.com/air-verse/air@latest

# 依存解決（初回ビルド短縮用。実運用は bind mount で上書きされる）
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# ソース（bind mount が無い場合でも動くように一旦コピー）
COPY . .

ENV CGO_ENABLED=0 \
    PORT=8080

# 40000 は delveのデバッグポート
EXPOSE 8080 40000

# ホットリロード起動
CMD ["air", "-c", ".air.toml"]
