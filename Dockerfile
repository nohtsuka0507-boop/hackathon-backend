# 1. ビルド環境
FROM golang:1.24-bookworm as builder

WORKDIR /app

# 依存関係の定義ファイルを先にコピー（キャッシュ活用のため）
COPY go.mod go.sum ./

# ライブラリをインターネットからダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# ビルド実行（-mod=vendor を削除し、通常ビルドにする）
RUN go build -v -o server .

# 2. 実行環境
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /server
ENV PORT 8080
CMD ["/server"]