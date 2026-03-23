import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useActions, type ActionAttribute } from "../../hooks/useActions";
import { useTags } from "../../hooks/useTags";
import { usePrototypes } from "../../hooks/usePrototypes";
import styles from "../../assets/styles/ActionForm.module.scss";

const MIN_TAG_COUNT = 1;
const MAX_TAG_COUNT = 30;
const MAX_TAG_NAME_LENGTH = 128;
const MAX_SUGGESTIONS = 8;
const RECENT_TAG_KEY = "action_form_recent_tags_v1";
const MAX_RECENT_TAGS = 200;

const normalizeTagName = (value: string) => value.trim().toLowerCase();
const ATTRIBUTE_TOKEN_RE = /^([^<>=\s]+)\s*(<=|>=|=|<|>)\s*(-?\d+(?:\.\d+)?)$/u;

const splitInputToTagNames = (input: string): string[] =>
  input
    .split(/[,\s]+/g)
    .map((v) => v.trim())
    .filter((v) => v.length > 0);

const mergeTagNames = (base: string[], incoming: string[]): string[] => {
  const next = [...base];
  const seen = new Set(base.map(normalizeTagName));
  for (const raw of incoming) {
    const cleaned = raw.trim();
    if (!cleaned) continue;
    const key = normalizeTagName(cleaned);
    if (seen.has(key)) continue;
    next.push(cleaned);
    seen.add(key);
    if (next.length >= MAX_TAG_COUNT) break;
  }
  return next;
};

const parseTagNamesAndAttributes = (names: string[]): { tagNames: string[]; attributes: ActionAttribute[] } => {
  const tagNames: string[] = [];
  const tagSeen = new Set<string>();
  const attributeMap = new Map<string, number>();

  for (const rawName of names) {
    const token = rawName.trim();
    if (!token) continue;

    const match = token.match(ATTRIBUTE_TOKEN_RE);
    if (match) {
      const key = match[1].trim().toLowerCase();
      const valueNumber = Number(match[3]);
      if (Number.isFinite(valueNumber)) {
        attributeMap.set(key, valueNumber);
        continue;
      }
    }

    const normalized = normalizeTagName(token);
    if (tagSeen.has(normalized)) continue;
    tagSeen.add(normalized);
    tagNames.push(token);
  }

  const attributes = Array.from(attributeMap.entries()).map(([key, value]) => ({
    key,
    value_number: value,
  }));

  return { tagNames, attributes };
};

const loadRecentTags = (): string[] => {
  try {
    const raw = localStorage.getItem(RECENT_TAG_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((v): v is string => typeof v === "string" && v.trim().length > 0);
  } catch {
    return [];
  }
};

const saveRecentTags = (tags: string[]) => {
  localStorage.setItem(RECENT_TAG_KEY, JSON.stringify(tags.slice(0, MAX_RECENT_TAGS)));
};

const updateRecentTags = (current: string[], usedNames: string[]): string[] => {
  const next: string[] = [];
  const seen = new Set<string>();
  for (const name of usedNames) {
    const key = normalizeTagName(name);
    if (!key || seen.has(key)) continue;
    next.push(name.trim());
    seen.add(key);
  }
  for (const name of current) {
    const key = normalizeTagName(name);
    if (!key || seen.has(key)) continue;
    next.push(name.trim());
    seen.add(key);
    if (next.length >= MAX_RECENT_TAGS) break;
  }
  return next;
};

const ActionForm = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = Boolean(id);

  const { loading, error, fetchActionByID, createAction, updateAction } = useActions();
  const { tags, fetchTags, createTag, loading: tagsLoading, error: tagsError } = useTags();
  const { prototypes, fetchPrototypes, loading: prototypesLoading, error: prototypesError } = usePrototypes();

  const [title, setTitle] = useState("");
  const [parentPrototypeID, setParentPrototypeID] = useState<number | null>(null);
  const [occurredAt, setOccurredAt] = useState<string>("");
  const [notes, setNotes] = useState("");
  const [selectedTagNames, setSelectedTagNames] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");
  const [activeSuggestionIndex, setActiveSuggestionIndex] = useState<number>(-1);
  const [recentTags, setRecentTags] = useState<string[]>([]);
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [initialTagIDs, setInitialTagIDs] = useState<number[] | null>(null);
  const [initialAttributeTokens, setInitialAttributeTokens] = useState<string[] | null>(null);

  useEffect(() => {
    fetchTags();
    fetchPrototypes();
    setRecentTags(loadRecentTags());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!isEditMode || !id) return;
    const load = async () => {
      const action = await fetchActionByID(Number(id));
      if (!action) {
        setFormError("行動記録が見つかりません。");
        return;
      }
      setTitle(action.title || "");
      setOccurredAt(action.occurred_at.slice(0, 16));
      setNotes(action.notes || "");
      setParentPrototypeID(null);
      setInitialTagIDs(action.tag_ids || []);
      const attrTokens =
        action.attributes
          ?.map((attr) =>
            typeof attr.value_number === "number" ? `${attr.key}=${attr.value_number}` : ""
          )
          .filter((v): v is string => v.length > 0) || [];
      setInitialAttributeTokens(attrTokens);
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isEditMode]);

  useEffect(() => {
    if (!initialTagIDs || tags.length === 0) return;
    const tagNameMap = new Map<number, string>();
    tags.forEach((tag) => tagNameMap.set(tag.id, tag.name));
    const names = initialTagIDs.map((tagID) => tagNameMap.get(tagID)).filter((v): v is string => Boolean(v));
    const merged = mergeTagNames(names, initialAttributeTokens || []);
    setSelectedTagNames(Array.from(new Set(merged)));
    setInitialTagIDs(null);
    setInitialAttributeTokens(null);
  }, [initialAttributeTokens, initialTagIDs, tags]);

  const prototypeMap = useMemo(() => {
    const map = new Map<number, (typeof prototypes)[number]>();
    prototypes.forEach((p) => map.set(p.id, p));
    return map;
  }, [prototypes]);

  const tagByNormalizedName = useMemo(() => {
    const map = new Map<string, (typeof tags)[number]>();
    tags.forEach((tag) => map.set(normalizeTagName(tag.name), tag));
    return map;
  }, [tags]);

  const recentIndexByKey = useMemo(() => {
    const map = new Map<string, number>();
    recentTags.forEach((name, idx) => {
      const key = normalizeTagName(name);
      if (!map.has(key)) map.set(key, idx);
    });
    return map;
  }, [recentTags]);

  const suggestions = useMemo(() => {
    const q = tagInput.trim().toLowerCase();
    if (!q) return [] as string[];

    const selectedSet = new Set(selectedTagNames.map(normalizeTagName));
    const candidates = tags
      .map((tag) => tag.name)
      .filter((name) => !selectedSet.has(normalizeTagName(name)))
      .filter((name) => name.toLowerCase().includes(q));

    const scored = candidates.map((name) => {
      const lower = name.toLowerCase();
      const startsWith = lower.startsWith(q);
      const idx = lower.indexOf(q);
      const recentIndex = recentIndexByKey.get(normalizeTagName(name));
      return {
        name,
        startsWithScore: startsWith ? 0 : 1,
        containsIndex: idx < 0 ? 9999 : idx,
        recentScore: recentIndex ?? 9999,
      };
    });

    scored.sort((a, b) => {
      if (a.startsWithScore !== b.startsWithScore) return a.startsWithScore - b.startsWithScore;
      if (a.recentScore !== b.recentScore) return a.recentScore - b.recentScore;
      if (a.containsIndex !== b.containsIndex) return a.containsIndex - b.containsIndex;
      return a.name.localeCompare(b.name);
    });

    return scored.slice(0, MAX_SUGGESTIONS).map((s) => s.name);
  }, [tagInput, selectedTagNames, tags, recentIndexByKey]);

  useEffect(() => {
    if (suggestions.length === 0) {
      setActiveSuggestionIndex(-1);
      return;
    }
    if (activeSuggestionIndex >= suggestions.length) {
      setActiveSuggestionIndex(0);
    }
  }, [suggestions, activeSuggestionIndex]);

  const pushRecent = (names: string[]) => {
    const next = updateRecentTags(recentTags, names);
    setRecentTags(next);
    saveRecentTags(next);
  };

  const addTagNames = (names: string[]) => {
    if (names.length === 0) return;
    setSelectedTagNames((prev) => mergeTagNames(prev, names));
    pushRecent(names);
  };

  const commitTagInput = () => {
    const names = splitInputToTagNames(tagInput);
    if (names.length > 0) addTagNames(names);
    setTagInput("");
  };

  const pickSuggestion = (name: string) => {
    addTagNames([name]);
    setTagInput("");
    setActiveSuggestionIndex(-1);
  };

  const removeTagName = (name: string) => {
    const targetKey = normalizeTagName(name);
    setSelectedTagNames((prev) => prev.filter((v) => normalizeTagName(v) !== targetKey));
  };

  const handleTagInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "ArrowDown") {
      if (suggestions.length === 0) return;
      e.preventDefault();
      setActiveSuggestionIndex((prev) => (prev < 0 ? 0 : (prev + 1) % suggestions.length));
      return;
    }
    if (e.key === "ArrowUp") {
      if (suggestions.length === 0) return;
      e.preventDefault();
      setActiveSuggestionIndex((prev) => (prev < 0 ? suggestions.length - 1 : (prev - 1 + suggestions.length) % suggestions.length));
      return;
    }
    if (e.key === "Escape") {
      setActiveSuggestionIndex(-1);
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      const hasSeparator = /[,\s]/.test(tagInput);
      if (!hasSeparator && suggestions.length > 0) {
        const idx = activeSuggestionIndex >= 0 ? activeSuggestionIndex : 0;
        pickSuggestion(suggestions[idx]);
        return;
      }
      commitTagInput();
      return;
    }
    if (e.key === "," || e.key === " ") {
      e.preventDefault();
      commitTagInput();
      return;
    }
    if (e.key === "Backspace" && tagInput.length === 0 && selectedTagNames.length > 0) {
      e.preventDefault();
      setSelectedTagNames((prev) => prev.slice(0, prev.length - 1));
    }
  };

  const handleParentPrototypeChange = (value: string) => {
    const nextID = Number(value);
    if (!nextID) {
      setParentPrototypeID(null);
      return;
    }
    setParentPrototypeID(nextID);
    const parent = prototypeMap.get(nextID);
    if (!parent || parent.tag_ids.length === 0) return;
    const parentNames = parent.tag_ids
      .map((tagID) => tags.find((tag) => tag.id === tagID)?.name)
      .filter((v): v is string => Boolean(v));
    setSelectedTagNames(Array.from(new Set(parentNames)));
    pushRecent(parentNames);
  };

  const resolveTagIDs = async (tagNames: string[]): Promise<number[] | null> => {
    const uniqueNames = Array.from(new Set(tagNames.map((v) => v.trim()).filter((v) => v.length > 0)));
    if (uniqueNames.length < MIN_TAG_COUNT) {
      setFormError(`タグは最低 ${MIN_TAG_COUNT} 件必要です。`);
      return null;
    }
    if (uniqueNames.length > MAX_TAG_COUNT) {
      setFormError(`タグは最大 ${MAX_TAG_COUNT} 件までです。`);
      return null;
    }

    const tagIDs: number[] = [];
    for (const name of uniqueNames) {
      if (name.length > MAX_TAG_NAME_LENGTH) {
        setFormError(`タグ名は ${MAX_TAG_NAME_LENGTH} 文字以内で入力してください。`);
        return null;
      }
      const key = normalizeTagName(name);
      const existing = tagByNormalizedName.get(key);
      if (existing) {
        tagIDs.push(existing.id);
        continue;
      }
      const created = await createTag(name);
      if (!created) {
        setFormError(`タグ作成に失敗しました: ${name}`);
        return null;
      }
      tagIDs.push(created.id);
      tagByNormalizedName.set(key, created);
    }
    return tagIDs;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!title.trim()) {
      setFormError("タイトルは必須です。");
      return;
    }
    if (title.length > 128) {
      setFormError("タイトルは128文字以内で入力してください。");
      return;
    }
    if (!occurredAt) {
      setFormError("発生日時は必須です。");
      return;
    }

    const finalTokens = mergeTagNames(selectedTagNames, splitInputToTagNames(tagInput));
    const parsedInput = parseTagNamesAndAttributes(finalTokens);

    setSubmitting(true);
    try {
      const tagIDs = await resolveTagIDs(parsedInput.tagNames);
      if (!tagIDs) return;

      const occurredAtISO = new Date(occurredAt).toISOString();
      const result =
        isEditMode && id
          ? await updateAction(
              Number(id),
              title.trim(),
              parentPrototypeID,
              occurredAtISO,
              notes.trim(),
              tagIDs,
              parsedInput.attributes
            )
          : await createAction(
              title.trim(),
              parentPrototypeID,
              occurredAtISO,
              notes.trim(),
              tagIDs,
              parsedInput.attributes
            );

      if (!result) {
        setFormError(isEditMode ? "行動記録の更新に失敗しました。" : "行動記録の作成に失敗しました。");
        return;
      }
      pushRecent(parsedInput.tagNames);
      navigate("/constructions/actions");
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "予期しないエラーが発生しました。");
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
      <h1 className={styles.title}>{isEditMode ? "行動記録を編集" : "行動記録を作成（= プロトタイプ）"}</h1>

      {(formError || error || tagsError || prototypesError) && (
        <div className={styles.errorBox}>エラー: {formError || error || tagsError || prototypesError}</div>
      )}

            <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="title">
            タイトル <span className={styles.required}>*</span>
          </label>
          <input
            id="title"
            className={styles.input}
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={128}
            placeholder="例: Aさん 新宿 13:00 山手線 乗車"
            disabled={submitting}
          />
          <small className={styles.hint}>{title.length}/128</small>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="parentPrototype">
            継承元プロトタイプ（任意）
          </label>
          <select
            id="parentPrototype"
            className={styles.select}
            value={parentPrototypeID ?? 0}
            onChange={(e) => handleParentPrototypeChange(e.target.value)}
            disabled={submitting || prototypesLoading}
          >
            <option value={0}>継承なし</option>
            {prototypes.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <small className={styles.hint}>継承元を選ぶと、親のタグ集合を初期値として取り込みます。</small>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="occurredAt">
            発生日時 <span className={styles.required}>*</span>
          </label>
          <input
            id="occurredAt"
            className={styles.input}
            type="datetime-local"
            value={occurredAt}
            onChange={(e) => setOccurredAt(e.target.value)}
            disabled={submitting}
          />
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="tagInput">
            タグ集合 <span className={styles.required}>*</span>
          </label>
          <div className={styles.tagInputArea}>
            <div className={styles.tagChips}>
              {selectedTagNames.map((name) => (
                <button
                  key={normalizeTagName(name)}
                  type="button"
                  className={styles.tagChip}
                  onClick={() => removeTagName(name)}
                  disabled={submitting}
                  title="削除"
                >
                  {name} x
                </button>
              ))}
            </div>
            <input
              id="tagInput"
              className={styles.input}
              type="text"
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={handleTagInputKeyDown}
              onBlur={commitTagInput}
              placeholder="タグを入力（Enter / Space / カンマで確定）"
              disabled={submitting}
            />
            {suggestions.length > 0 && (
              <div className={styles.suggestionList}>
                {suggestions.map((name, idx) => (
                  <button
                    key={normalizeTagName(name)}
                    type="button"
                    className={`${styles.suggestionItem} ${idx === activeSuggestionIndex ? styles.suggestionItemActive : ""}`}
                    onMouseEnter={() => setActiveSuggestionIndex(idx)}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      pickSuggestion(name);
                    }}
                    disabled={submitting}
                  >
                    {name}
                  </button>
                ))}
              </div>
            )}
          </div>
          <small className={styles.hint}>
            {"最低 "}
            {MIN_TAG_COUNT}
            {" 件、最大 "}
            {MAX_TAG_COUNT}
            {" 件。未登録タグは保存時に作成されます。`年齢=32` / `運賃>=180` / `duration<15` のような `key 演算子 数値` は自動で属性として保存されます。"}
          </small>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="notes">
            メモ
          </label>
          <textarea
            id="notes"
            className={styles.textarea}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={5}
            placeholder="メモ"
            disabled={submitting}
          />
        </div>

        <div className={styles.actions}>
          <button className={styles.primaryButton} type="submit" disabled={submitting || tagsLoading || prototypesLoading}>
            {submitting ? "保存中..." : isEditMode ? "更新" : "作成"}
          </button>
          <button className={styles.secondaryButton} type="button" onClick={() => navigate("/constructions/actions")} disabled={submitting}>
            キャンセル
          </button>
        </div>
      </form>
    </div>
  );
};

export default ActionForm;



