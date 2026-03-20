import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTargets } from "../hooks/useTargets";
import { useTags } from "../hooks/useTags";
import styles from "../assets/styles/TargetForm.module.scss";

const TargetForm = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = Boolean(id);

  const { loading, error, fetchTargetByID, createTarget, updateTarget } = useTargets();
  const { tags, fetchTags, loading: tagsLoading, error: tagsError } = useTags();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [selectedTagIDs, setSelectedTagIDs] = useState<number[]>([]);
  const [queryText, setQueryText] = useState("");
  const [anyTagIDs, setAnyTagIDs] = useState<number[]>([]);
  const [anyTagGroups, setAnyTagGroups] = useState<number[][]>([]);
  const [excludeTagIDs, setExcludeTagIDs] = useState<number[]>([]);
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchTags();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!isEditMode || !id) {
      return;
    }
    const load = async () => {
      const target = await fetchTargetByID(Number(id));
      if (!target) {
        setFormError("観察対象が見つかりません。");
        return;
      }
      setName(target.name);
      setDescription(target.description || "");
      setSelectedTagIDs(target.tag_ids || []);
      setQueryText(target.query_text || "");
      setAnyTagIDs(target.any_tag_ids || []);
      setAnyTagGroups(target.any_tag_groups || []);
      setExcludeTagIDs(target.exclude_tag_ids || []);
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isEditMode]);

  const groupedTags = useMemo(() => {
    const groups = new Map<string, typeof tags>();
    tags.forEach((tag) => {
      const key = tag.group_name || "未分類";
      if (!groups.has(key)) {
        groups.set(key, []);
      }
      groups.get(key)?.push(tag);
    });
    return Array.from(groups.entries());
  }, [tags]);

  const toggleTag = (tagID: number) => {
    setSelectedTagIDs((prev) => (prev.includes(tagID) ? prev.filter((id) => id !== tagID) : [...prev, tagID]));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!name.trim()) {
      setFormError("名前は必須です。");
      return;
    }
    if (name.length > 64) {
      setFormError("名前は64文字以内で入力してください。");
      return;
    }
    if (selectedTagIDs.length === 0) {
      setFormError("タグを1件以上選択してください。");
      return;
    }

    setSubmitting(true);
    try {
      const result = isEditMode && id
        ? await updateTarget(Number(id), name.trim(), description.trim(), selectedTagIDs, {
            queryText,
            anyTagIDs,
            anyTagGroups,
            excludeTagIDs,
          })
        : await createTarget(name.trim(), description.trim(), selectedTagIDs, {
            queryText,
            anyTagIDs,
            anyTagGroups,
            excludeTagIDs,
          });

      if (!result) {
        setFormError(isEditMode ? "更新に失敗しました。" : "作成に失敗しました。");
        return;
      }
      navigate("/targets");
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "エラーが発生しました。");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading && isEditMode) {
    return (
      <div className={styles.loading}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>{isEditMode ? "観察対象を編集" : "観察対象を作成"}</h1>

      {(formError || error || tagsError) && (
        <div className={styles.errorBox}>エラー: {formError || error || tagsError}</div>
      )}

      <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="name">
            名前 <span className={styles.required}>*</span>
          </label>
          <input
            className={styles.input}
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={64}
            required
            placeholder="例: 平日昼の通勤"
            disabled={submitting}
          />
          <small className={styles.hint}>{name.length}/64</small>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="description">
            説明
          </label>
          <textarea
            className={styles.textarea}
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="この観察対象の意図や使いどころを入力"
            disabled={submitting}
          />
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label}>
            タグ条件（AND） <span className={styles.required}>*</span>
          </label>
          {tagsLoading ? (
            <p>タグを読み込み中...</p>
          ) : tags.length === 0 ? (
            <p>タグがありません。先にタグを作成してください。</p>
          ) : (
            <div className={styles.tagSelector}>
              {groupedTags.map(([groupName, groupTags]) => (
                <div key={groupName} className={styles.tagGroup}>
                  <div className={styles.tagGroupTitle}>{groupName}</div>
                  <div className={styles.tagList}>
                    {groupTags.map((tag) => {
                      const checked = selectedTagIDs.includes(tag.id);
                      return (
                        <label key={tag.id} className={styles.tagItem}>
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => toggleTag(tag.id)}
                            disabled={submitting}
                          />
                          <span>{tag.name}</span>
                        </label>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
          <small className={styles.hint}>選択したタグすべてを含む行動記録が一致対象になります。</small>
        </div>

        <div className={styles.actions}>
          <button className={styles.primaryButton} type="submit" disabled={submitting || tagsLoading}>
            {submitting ? "送信中..." : isEditMode ? "更新" : "作成"}
          </button>
          <button className={styles.secondaryButton} type="button" onClick={() => navigate("/targets")} disabled={submitting}>
            キャンセル
          </button>
        </div>
      </form>
    </div>
  );
};

export default TargetForm;
