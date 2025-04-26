# # ビルド用のステージ（マルチステージビルドを使用）
# FROM golang:1.20 AS builder

# # 作業ディレクトリを指定
# WORKDIR /server/app

# # ホストのファイルをコンテナにコピー
# COPY . .

# # モジュールのダウンロード（不要なものを削除し、必要なものを取得）
# RUN go mod tidy

# # ビルドし、実行ファイル「main」を生成
# RUN go build -o main .


# # 実行用のステージ
# FROM alpine:latest

# # 作業ディレクトリを指定
# WORKDIR /root/

# # 「builder」ステージで生成した実行ファイル「main」をコピー
# COPY --from=builder /server/app/main .

# # コンテナ起動時に「main」を実行
# CMD ["./main"]
