#  ビルドステージ
FROM golang:1.22 as builder

# プロジェクトルートを作業ディレクトリに設定
WORKDIR /server

# go.modとgo.sumを先にコピーして依存取得（キャッシュを活用）
COPY go.mod go.sum ./
RUN go mod download

# 全体をコンテナにコピー（ルート構造維持）
COPY . .

# main.go をビルドしてバイナリを作成
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main/cmd/main.go

#  実行ステージ
FROM alpine:latest

# 最小限の実行環境（必要であれば tzdata 等を追加）
RUN apk --no-cache add tzdata

# 作業ディレクトリを統一して /app に設定
WORKDIR /server

# ビルド成果物（app バイナリとすべての構造）をコピー
COPY --from=builder /server /server

# アプリケーションを実行
CMD ["./server"]
