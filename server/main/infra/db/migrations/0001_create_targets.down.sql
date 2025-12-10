-- マイグレーションのロールバック時にtargetsテーブルとそのインデックスを削除する
DROP INDEX IF EXISTS idx_targets_not_deleted;
DROP TABLE IF EXISTS targets;
