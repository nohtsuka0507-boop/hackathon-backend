# 1. ビルド用のGo環境を準備
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 必要なモジュールを先にダウンロード
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションのコードをコピー
COPY . .

# アプリケーションをビルド（コンパイル）
# CGO_ENABLED=0 は、GoがC言語のコードに依存しないようにするおまじない
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

# 2. 実行用の小さな環境を準備
FROM gcr.io/distroless/base-debian11

WORKDIR /

# 1. でビルドした "server" プログラムだけをコピー
COPY --from=builder /app/server .

# このコンテナが起動したときに実行するコマンド
# Cloud Runが "8000" ポートに来るように設定されている
ENV PORT 8000
CMD ["/server"]