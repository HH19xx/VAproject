-- マイグレーションのロールバック時にtargetsテーブルとそのインデックスを削除する
DROP INDEX IF EXISTS unique_active_target_name;
DROP TABLE IF EXISTS targets;
