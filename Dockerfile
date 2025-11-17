# 1. ビルド用のGo環境を準備 (Go 1.25)
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 必要なモジュールを先にダウンロード/整理
COPY go.mod go.sum ./
RUN go mod tidy

# アプリケーションのコードをコピー
COPY . .

# アプリケーションをビルド（コンパイル）
# CGO_ENABLED=0 は静的リンクのため、GOOS=linux はLinux実行ファイルのため
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

# --- ↓ここからが変更点（distrolessの代わり）↓ ---

# 2. 実行用の「小さなAlpine環境」を準備
FROM alpine:3.18

WORKDIR /app

# 1. でビルドした "server" プログラムだけをコピー
COPY --from=builder /app/server .

# このコンテナが起動したときに実行するコマンド
# Cloud Runが "8000" ポートに来るように設定されている
ENV PORT 8000
CMD ["/app/server"]