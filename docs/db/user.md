# user テーブルの定義

ユーザー情報を管理するためのテーブルです。

## フィールド定義

| フィールド名   | データ型      | 制約                     | 説明                       |
|--------------|--------------|--------------------------|----------------------------|
| id           | INT          | PRIMARY KEY AUTO_INCREMENT | ユーザーの一意な識別子（自動採番）|
| username     | VARCHAR(32)  | NOT NULL UNIQUE          | ユーザー名（最大32文字）   |
| email        | VARCHAR(128) | NOT NULL UNIQUE          | ユーザーのメールアドレス（最大128文字）|
| password_hash| VARCHAR(128) | NOT NULL                 | パスワードのハッシュ値（最大128文字）|
| created_at   | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時          |
| create_user  | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| updated_at   | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user  | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_user                     | id                  | PRIMARY KEY|
| UQ_user_username            | username            | UNIQUE     |
| UQ_user_email               | email               | UNIQUE     |