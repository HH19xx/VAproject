# temp_actions テーブルの定義

## 用途
ユーザーが入力途中の行動データを保存し、後で確定処理を実施するためのテーブルです。

## テーブル構成

| フィールド名   | データ型      | 制約                     | 説明                           |
|----------------|--------------|--------------------------|--------------------------------|
| id             | INT          | PRIMARY KEY AUTO_INCREMENT | 一意な識別子（自動採番）      |
| user_id        | INT          | INDEX, FOREIGN KEY REFERENCES user(id) | ユーザーのID                   |
| action_type_id | INT          | INDEX, FOREIGN KEY REFERENCES action_types(id) | 行動の種類（マスタ参照）      |
| target_id      | INT          | INDEX, FOREIGN KEY REFERENCES target(id) | 観察対象のID                   |
| created_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時               |
| create_user    | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| notes          | TEXT         |                          | 任意の補足情報                 |
| updated_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user    | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_temp_actions             | id                  | PRIMARY KEY|
| IDX_temp_actions_user       | user_id             | INDEX      |
| IDX_temp_actions_action_type| action_type_id      | INDEX      |
| IDX_temp_actions_target     | target_id           | INDEX      |