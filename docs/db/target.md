# target テーブルの定義

## 概要
`target` テーブルは、観察対象の情報を管理するためのテーブルです。このテーブルは、観察対象の属性を格納します。

## テーブル定義

| フィールド名   | データ型      | 制約                     | 説明                           |
|----------------|--------------|--------------------------|--------------------------------|
| id             | INT          | PRIMARY KEY AUTO_INCREMENT | 観察対象の一意な識別子（自動採番）|
| name           | VARCHAR(64)  | NOT NULL UNIQUE          | 観察対象の名前（最大64文字）   |
| description    | TEXT         |                          | 観察対象の説明（任意）        |
| created_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時              |
| create_user    | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| updated_at     | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user    | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## 使用例
- `id`: 1
- `name`: "サンプルターゲット"
- `description`: "これはサンプルの観察対象です。"
- `created_at`: 2023-10-01 12:00:00

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_target                   | id                  | PRIMARY KEY|
| UQ_target_name              | name                | UNIQUE     |