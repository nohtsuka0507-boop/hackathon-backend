FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/server .
ENV PORT 8000
CMD ["/app/server"]
