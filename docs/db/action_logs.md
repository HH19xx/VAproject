# action_logs テーブル定義

## 概要
`action_logs` テーブルは、観察対象の行動履歴を記録するためのテーブルです。このテーブルは、行動の傾向分析に利用されます。

## テーブル構成

| フィールド名   | データ型   | 制約                     | 説明                           |
|----------------|------------|--------------------------|--------------------------------|
| id             | INT        | PRIMARY KEY AUTO_INCREMENT | ログの一意な識別子（自動採番）|
| target_id      | INT        | INDEX, FOREIGN KEY REFERENCES target(id) | 観察対象のID                  |
| action_type    | INT        | INDEX, FOREIGN KEY REFERENCES action_types(id) | 行動の種類（マスタ参照）      |
| timestamp      | TIMESTAMP  | NOT NULL, INDEX          | 行動が発生した日時            |
| notes          | TEXT       |                          | 任意の補足情報                |
| created_at     | TIMESTAMP  | DEFAULT CURRENT_TIMESTAMP | レコード作成日時              |
| create_user    | VARCHAR(32)| NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| updated_at     | TIMESTAMP  | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user    | VARCHAR(32)| NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## 使用例
- 行動履歴を記録することで、ユーザーの行動パターンを分析し、システムの改善に役立てることができます。

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_action_logs              | id                  | PRIMARY KEY|
| IDX_action_logs_target      | target_id           | INDEX      |
| IDX_action_logs_action_type | action_type         | INDEX      |
| IDX_action_logs_timestamp   | timestamp           | INDEX      |