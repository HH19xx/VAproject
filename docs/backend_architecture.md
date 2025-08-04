# バックエンドのディレクトリ構成

VAProject
- server/
  - main/                 # 認証・ユーザー周りのメインサービス
    - cmd/              # 起動（main.go）
    - app/              # ユースケース・DTO
      - usecase/
      - dto/
    - domain/           # モデル・リポジトリIF
    - infra/            # DB・API等の実装
      - db/
      - repository/
    - interface/        # HTTP・ミドルウェア
      - handler/
      - middleware/

## アプリごとの詳細構成

1. アカウント登録・統合アプリ
main/
- cmd/
  - main.go
- app/
  - usecase/
    - user_usecase.go
  - dto/
    - user_dto.go
- domain/
  - user.go
- infra/
  - db/
  - repository/
    - user_repository.go
- interface/
  - handler/
    - user_handler.go
  - middleware/
    - auth_middleware.go
特徴: 新規ユーザーの登録機能や他2つのアプリとの統合ロジックを実装。

2. 観察対象アプリ (/target)  # 後で追加、今は作らない  
target/
- cmd/
  - main.go
- app/
  - usecase/
    - target_usecase.go
  - dto/
    - target_dto.go
- domain/
  - target.go
- infra/
  - db/
  - repository/
    - target_repository.go
- interface/
  - handler/
    - target_handler.go
  - middleware/
    - auth_middleware.go
特徴: 管理者向けに特化したエンドポイントや機能（統計データ管理、ユーザー管理など）を実装。

3. 行動記録アプリ (/action)  # 後で追加、今は作らない  
action/
- cmd/
  - main.go
- app/
  - usecase/
    - action_usecase.go
  - dto/
    - action_dto.go
- domain/
  - action.go
- infra/
  - db/
  - repository/
    - action_repository.go
- interface/
  - handler/
    - action_handler.go
  - middleware/
    - auth_middleware.go
特徴: 顧客が利用する機能（データの閲覧やダッシュボード表示など）を中心に構築。

4. 行動種類登録アプリ (customer/)  # 後で追加、今は作らない  
action_type/
- cmd/
  - main.go
- app/
  - usecase/
    - action_type_usecase.go
  - dto/
    - action_type_dto.go
- domain/
  - action_type.go
- infra/
  - db/
  - repository/
    - action_type_repository.go
- interface/
  - handler/
    - action_type_handler.go
  - middleware/
    - auth_middleware.go

5. 統計アプリ (customer/)  # 後で追加、今は作らない  
statistics/
- cmd/
  - main.go
- app/
  - usecase/
    - statistics_usecase.go
  - dto/
    - statistics_dto.go
- domain/
  - statistics.go
- infra/
  - db/
  - repository/
    - statistics_repository.go
- interface/
  - handler/
    - statistics_handler.go
  - middleware/
    - auth_middleware.go

## ポイント

1. マイクロサービスアーキテクチャとDDDを採用し、独立性と共通性のバランスをとる。  
各アプリは独自のルーティングやエンドポイントを持つ。  

2. 依存関係の管理  
各アプリ (apps/) は直接他のアプリに依存しない。  

3. 統合アプリの役割  
アカウント登録や、他アプリケーションへの連携を担当する。  
