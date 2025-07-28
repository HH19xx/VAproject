# action_types テーブル定義

このファイルには、`action_types` テーブルの定義が含まれています。このテーブルは、行動の種類を定義するために使用されます。

## テーブル名: action_types

| フィールド名   | データ型      | 制約                     | 説明                     |
|----------------|--------------|--------------------------|--------------------------|
| id             | INT          | PRIMARY KEY AUTO_INCREMENT | 行動の一意な識別子（自動採番）|
| action_name    | VARCHAR(64)  | NOT NULL UNIQUE          | 行動の種類名（最大64文字）|
| description    | TEXT         |                          | 行動の説明（任意）      |
| created_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時        |
| create_user    | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| updated_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user    | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## 例

| id | action_name | description |
|----|-------------|-------------|
| 1  | 移動        | 物理的な移動を示す行動 |
| 2  | 飲食        | 食事を取る行動         |

## インデックス定義

| インデックス名         | 対象カラム      | 種類         |
|----------------------|----------------|------------|
| PK_action_types      | id             | PRIMARY KEY|
| UQ_action_name       | action_name    | UNIQUE     |