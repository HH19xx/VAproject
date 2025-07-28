# settings テーブル定義

## テーブル名: settings

### 用途
システム設定値を管理します。

### 構成
| フィールド名 | データ型      | 制約          | 説明                     |
|--------------|--------------|---------------|--------------------------|
| id           | INT          | PRIMARY KEY AUTO_INCREMENT | 設定の一意な識別子（自動採番）|
| key          | VARCHAR(64)  | NOT NULL UNIQUE| 設定のキー（最大64文字） |
| value        | VARCHAR(256) | NOT NULL      | 設定の値（最大256文字）  |
| created_at   | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP | レコード作成日時        |
| create_user  | VARCHAR(32)  | NOT NULL                 | 作成者ユーザーID/名（最大32文字）|
| updated_at   | TIMESTAMP    | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | レコード最終更新日時 |
| update_user  | VARCHAR(32)  | NOT NULL                 | 更新者ユーザーID/名（最大32文字）|

## インデックス定義

| インデックス名                | 対象カラム           | 種類         |
|-----------------------------|---------------------|------------|
| PK_settings                 | id                  | PRIMARY KEY|
| UQ_settings_key             | key                 | UNIQUE     |

### 備考
- キーバリュー形式で便利に使用されます。