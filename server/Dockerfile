# server/Dockerfile

# ビルドステージ
FROM golang:1.22 as builder

# 作業ディレクトリを /app に設定
WORKDIR /app

# ホストの server/ ディレクトリ内容をコンテナの /app にコピー
COPY . .

# 必要なモジュールをダウンロード
RUN go mod download

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app ./apps/registration/main.go

# 実行ステージ
FROM alpine:latest

# タイムゾーンを設定 (必要に応じて)
RUN apk --no-cache add tzdata

# 作業ディレクトリを /app_data に設定
WORKDIR /app_data

# ビルドステージから実行ファイルをコピー
COPY --from=builder /app/app /app_data/app

# マイグレーションファイルをコピー (ワーキングディレクトリ基準のパスに合わせる)
COPY --from=builder /app/infra/db/migrations /app_data/infra/db/migrations

# アプリケーションを起動
CMD ["./app"]
