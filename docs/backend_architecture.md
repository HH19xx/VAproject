バックエンドのディレクトリ構成
server/
├── apps/
│   ├── admin/                # 管理者向けアプリ
│   ├── customer/             # 顧客向けアプリ
│   └── registration/         # アカウント登録兼統合アプリ
├── core/                     # 共通ロジック (ユースケースやエンティティ)
│   ├── domain/               # エンティティとドメインモデル
│   ├── usecases/             # ユースケース (アプリケーション層)
│   └── shared/               # 共通ユーティリティやDTO
├── infra/                    # インフラ層 (データベース, 外部API)
│   ├── db/                   # データベースアクセス
│   ├── api/                  # 外部サービスとの接続
│   └── repository/           # リポジトリの実装
├── interface/                # プレゼンテーション層
│   ├── web/                  # HTTPハンドラ (APIエンドポイント)
│   ├── cli/                  # CLI用のインターフェース (必要なら)
│   └── middleware/           # 共通ミドルウェア
└── config/                   # 設定ファイル (環境変数, 設定値)

アプリごとの詳細構成
1. 管理者向けアプリ (apps/admin)
apps/admin/
├── main.go                   # エントリーポイント
├── handler/                  # プレゼンテーション層 (管理者APIエンドポイント)
│   └── admin_handler.go
├── routes/                   # ルーティング設定
│   └── admin_routes.go
├── service/                  # アプリケーション固有のビジネスロジック
│   └── admin_service.go
└── repository/               # データアクセス (インフラ層)
    └── admin_repository.go
特徴: 管理者向けに特化したエンドポイントや機能（統計データ管理、ユーザー管理など）を実装。

2. 顧客向けアプリ (apps/customer)
apps/customer/
├── main.go                   # エントリーポイント
├── handler/                  # プレゼンテーション層 (顧客APIエンドポイント)
│   └── customer_handler.go
├── routes/                   # ルーティング設定
│   └── customer_routes.go
├── service/                  # アプリケーション固有のビジネスロジック
│   └── customer_service.go
└── repository/               # データアクセス (インフラ層)
    └── customer_repository.go
特徴: 顧客が利用する機能（データの閲覧やダッシュボード表示など）を中心に構築。

3. アカウント登録・統合アプリ (apps/registration)
apps/registration/
├── main.go                   # エントリーポイント
├── handler/                  # プレゼンテーション層 (登録と統合API)
│   └── registration_handler.go
├── routes/                   # ルーティング設定
│   └── registration_routes.go
├── service/                  # アプリケーション固有のビジネスロジック
│   └── registration_service.go
└── repository/               # データアクセス (インフラ層)
    └── registration_repository.go
特徴: 新規ユーザーの登録機能や他2つのアプリとの統合ロジックを実装。

共通部分 (core/, infra/, interface/)
1. core/ (ドメインロジック)
core/
├── domain/
│   ├── user.go               # ユーザーモデル
│   ├── admin.go              # 管理者モデル
│   └── statistics.go         # 統計データモデル
├── usecases/
│   ├── user_usecase.go       # ユーザー関連のユースケース
│   ├── admin_usecase.go      # 管理者関連のユースケース
│   └── statistics_usecase.go # 統計データ関連のユースケース
└── shared/
    ├── dto/                  # 共通DTO
    │   └── user_dto.go
    ├── error.go              # 共通エラー定義
    └── utils.go              # ユーティリティ関数
目的: ドメインロジックとアプリケーションロジックを一箇所に集約し、再利用性を高める。

2. infra/ (インフラ層)
infra/
├── db/
│   ├── connection.go         # データベース接続設定
│   └── migrations/           # マイグレーションファイル
├── api/
│   └── external_service.go   # 外部サービスのクライアント
└── repository/
    ├── user_repository.go    # ユーザーデータアクセス
    ├── admin_repository.go   # 管理者データアクセス
    └── statistics_repository.go # 統計データアクセス

3. interface/ (プレゼンテーション層)
interface/
├── web/
│   ├── middleware.go         # 認証やエラーハンドリングのミドルウェア
│   ├── response.go           # 統一的なレスポンスフォーマット
│   └── routes.go             # ルート定義
├── cli/
│   └── main.go               # CLIインターフェース (必要なら)
└── middleware/
    ├── auth_middleware.go    # 認証ミドルウェア
    └── logging_middleware.go # ロギングミドルウェア


ポイント
1. クリーンアーキテクチャを採用し、独立性と共通性のバランスをとる。
各アプリは独自のルーティングやエンドポイントを持ちつつ、共通のドメインロジックやユースケースを core/ に集約して再利用する。

2. 依存関係の管理
各アプリ (apps/) は、core/ や infra/ を依存する形で動作し、直接他のアプリに依存しない設計。

3. 統合アプリの役割
アカウント登録や、他アプリケーションへの統合を担当する。
