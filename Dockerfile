# 1. ビルド環境 (Go 1.24 を指定して環境を作ります)
FROM golang:1.24-bookworm as builder

# 作業ディレクトリを作成
WORKDIR /app

# 依存関係ファイルをコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
# 念のためビルド中に整理
RUN go mod tidy
RUN go build -v -o server .

# 2. 実行環境 (軽量なLinuxを使います)
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# ビルドしたアプリをコピー
COPY --from=builder /app/server /server

# ポート8080を開放する設定
ENV PORT 8080

# アプリを起動
CMD ["/server"]