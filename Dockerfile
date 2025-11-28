# 1. ビルド用環境 (Goのコンパイルを行う)
FROM golang:1.25-alpine AS builder

# 作業ディレクトリを作成
WORKDIR /app

# 依存関係ファイルをコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードを全てコピー
COPY . .

RUN go mod tidy
# バイナリファイル（実行ファイル）をビルド
# CGO_ENABLED=0 は軽量なAlpine Linuxで動かすために必須
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 2. 実行用環境 (完成したバイナリだけを乗せる軽量環境)
FROM alpine:latest

# セキュリティ証明書を入れる（外部API通信に必要）
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルド環境から実行ファイルだけをコピー
COPY --from=builder /app/main .

# ポート8080を開ける
EXPOSE 8080

# サーバー起動
CMD ["./main"]