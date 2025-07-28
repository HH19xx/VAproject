# error_logs テーブル定義

## 概要
`error_logs` テーブルは、システムエラーや例外を記録するためのテーブルです。このテーブルは、トラブルシューティングやシステムの健全性を監視するために使用されます。

## テーブル構成

| フィールド名   | データ型      | 制約                     | 説明                           |
|----------------|--------------|--------------------------|--------------------------------|
| id             | INT          | PRIMARY KEY AUTO_INCREMENT | ログの一意な識別子（自動採番）|
| error_message  | TEXT         | NOT NULL                 | エラーメッセージ              |
| error_code     | VARCHAR(32)  |                          | エラーコード（最大32文字、任意）|
| created_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時              |
| create_user    | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| stack_trace    | TEXT         |                          | スタックトレース（任意）      |
| user_id        | INT          | INDEX, FOREIGN KEY REFERENCES user(id) | エラー発生時のユーザーID（任意）|
| updated_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user    | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_error_logs               | id                  | PRIMARY KEY|
| IDX_error_logs_user         | user_id             | INDEX      |