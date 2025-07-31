# 進捗

VAproject
├── client/ (フロントエンドのルートディレクトリ)
│   ├── README.md
│   ├── eslint.config.js
│   ├── index.html
│   ├── package-lock.json
│   ├── package.json
│   ├── public/
│   │   └── vite.svg
│   ├── src/
│   │   ├── App.css
│   │   ├── App.tsx
│   │   ├── assets/
│   │   │   └── react.svg
│   │   ├── components/
│   │   │   └── LoginForm.tsx
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   └── useFetchMessages.ts
│   │   ├── index.css
│   │   ├── main.tsx
│   │   ├── styles/
│   │   └── vite-env.d.ts
│   ├── tsconfig.app.json
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── vite.config.ts
├── docs/
│   ├── backend_architecture.md
│   ├── database_architecture.md
│   ├── db/
│   │   ├── README.md
│   │   ├── action_logs.md
│   │   ├── action_types.md
│   │   ├── error_logs.md
│   │   ├── settings.md
│   │   ├── target.md
│   │   ├── temp_actions.md
│   │   └── user.md
│   ├── frontend_architecture.md
│   ├── progress.md
│   └── schedule.md
└── server/ (バックエンドのルートディレクトリ)
    ├── Dockerfile
    ├── docker-compose.yml
    ├── go.mod
    ├── go.sum
    ├── infra/
    │   └── db/
    │       └── init/
    └── main/
        ├── app/
        │   └── usecases/
        │       └── user_usecase.go
        ├── cmd/
        │   └── main.go
        ├── domain/
        │   └── user.go
        ├── infra/
        │   ├── db/
        │   │   ├── connection.go
        │   │   ├── connection_test.go
        │   │   ├── init/
        │   │   ├── migrate.go
        │   │   └── migrations/
        │   │       ├── 0001_create_action_logs.up.sql
        │   │       ├── 0001_create_action_types.sql
        │   │       ├── 0001_create_custom_actions.up.sql
        │   │       ├── 0001_create_error_logs.up.sql
        │   │       ├── 0001_create_settings.up.sql
        │   │       ├── 0001_create_targets.up.sql
        │   │       ├── 0001_create_temp_actions.up.sql
        │   │       ├── 0001_create_users.down.sql
        │   │       ├── 0001_create_users.up.sql
        │   │       ├── 0002_add_email_to_users.down.sql
        │   │       └── 0002_add_email_to_users.up.sql
        │   ├── jwt/
        │   │   └── jwt_utils.go
        │   ├── oauth/
        │   │   └── oauth_config.go
        │   └── repository/
        │       └── user_repository.go
        └── interface/
            ├── handler/
            │   ├── auth_handler.go
            │   ├── main_handler.go
            │   └── oauth_handler.go
            └── middleware/
                └── auth_middleware.go

## 実装済みの機能

### バックエンド
- [x] 基本的なGo構成 (Echo未実装)
- [x] データベース接続 (PostgreSQL)
- [x] マイグレーション機能
- [x] JWT認証機能
- [x] OAuth (Google) 実装準備
- [x] ユーザー管理基本機能
- [x] CORS設定

### フロントエンド
- [x] React + Vite + TypeScript環境
- [x] 認証コンテキスト
- [x] ログインフォーム
- [x] 基本的なAPI通信

### データベース
- [x] ユーザーテーブル
- [x] 各種テーブル (action_logs, action_types, targets等)
- [x] マイグレーション設定

## 今後の実装予定

- [ ] OAuth実装の完了
- [ ] CRUD機能の実装
- [ ] フロントエンドのUI/UX改善
- [ ] 統計機能の実装
- [ ] テスト実装