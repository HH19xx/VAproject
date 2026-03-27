import SearchSuggestionRail from "../../components/search/SearchSuggestionRail";
import styles from "../../assets/styles/Actions.module.scss";
import type { SearchTagSuggestion, SortKey, SortOrder } from "./actionsSearch/actionsSearchSectionHelpers";

interface ActionsSearchSectionProps {
  sort: SortKey;
  order: SortOrder;
  from: string;
  to: string;
  isDateRangeManual: boolean;
  danbooruQuery: string;
  searchSuggestions: SearchTagSuggestion[];
  groupedTags: Array<[string, Array<{ id: number; name: string }>]>; 
  filterTagIDs: number[];
  loading: boolean;
  snapshotLoading: boolean;
  targetSaving: boolean;
  onSortChange: (value: SortKey) => void;
  onOrderChange: (value: SortOrder) => void;
  onFromChange: (value: string) => void;
  onToChange: (value: string) => void;
  onDanbooruQueryChange: (value: string) => void;
  onAppendSearchSuggestion: (tagName: string) => void;
  onResetDateRangeSync: () => void;
  onToggleFilterTag: (tagID: number) => void;
  onSearch: () => void | Promise<void>;
  onSaveAsTarget: () => void | Promise<void>;
  onClear: () => void | Promise<void>;
}

const ActionsSearchSection = ({
  sort,
  order,
  from,
  to,
  isDateRangeManual,
  danbooruQuery,
  searchSuggestions,
  groupedTags,
  filterTagIDs,
  loading,
  snapshotLoading,
  targetSaving,
  onSortChange,
  onOrderChange,
  onFromChange,
  onToChange,
  onDanbooruQueryChange,
  onAppendSearchSuggestion,
  onResetDateRangeSync,
  onToggleFilterTag,
  onSearch,
  onSaveAsTarget,
  onClear,
}: ActionsSearchSectionProps) => {
  const hasQuery = danbooruQuery.trim().length > 0;

  return (
    <div className={styles.filterPanel}>
      <div className={styles.searchHero}>
        <div className={styles.filterTitle}>検索条件</div>
      </div>

      <div className={styles.controlGrid}>
        <label>
          並び順
          <select className={styles.select} value={sort} onChange={(event) => onSortChange(event.target.value as SortKey)}>
            <option value="occurred_at">発生日時</option>
            <option value="created_at">作成日時</option>
            <option value="updated_at">更新日時</option>
            <option value="title">タイトル</option>
            <option value="tag_count">タグ数</option>
          </select>
        </label>
        <label>
          順序
          <select className={styles.select} value={order} onChange={(event) => onOrderChange(event.target.value as SortOrder)}>
            <option value="desc">降順</option>
            <option value="asc">昇順</option>
          </select>
        </label>
        <label>
          期間開始
          <input className={styles.input} type="datetime-local" value={from} onChange={(event) => onFromChange(event.target.value)} />
        </label>
        <label>
          期間終了
          <input className={styles.input} type="datetime-local" value={to} onChange={(event) => onToChange(event.target.value)} />
        </label>
      </div>

      <div className={styles.info}>
        期間設定: {isDateRangeManual ? "手動指定中" : "外部データ設定と自動同期中"}
        {isDateRangeManual && (
          <>
            {" "}
            <button type="button" className={styles.suggestionChip} onClick={onResetDateRangeSync}>
              自動同期へ戻す
            </button>
          </>
        )}
      </div>

      <div className={styles.searchBarRow}>
        <label className={styles.searchBarField}>
          Danbooru 形式検索クエリ
          <input
            className={`${styles.input} ${styles.searchBarInput}`}
            type="text"
            value={danbooruQuery}
            onChange={(event) => onDanbooruQueryChange(event.target.value)}
            placeholder="例: 交通 山手線 -混雑 / 昼_弁当 / 価格=680 / 交通 山手線 | 物価 昼_弁当"
          />
        </label>
      </div>

      <SearchSuggestionRail
        title={hasQuery ? "入力候補" : "候補タグ"}
        suggestions={searchSuggestions}
        emptyMessage={hasQuery ? "一致するタグはありません。" : "候補に出せるタグがまだありません。"}
        onSelectSuggestion={onAppendSearchSuggestion}
      />

      <div className={styles.info}>
        半角空白区切りです。`tag` は AND、`~tag` は OR、`-tag` は除外です。`A B | C D` は `(A B)~(C D)`、つまり `(A AND B) OR (C AND D)` を意味します。タグ名に半角スペースを入れたい場合は `_` を使ってください。
      </div>

      <details className={styles.collapsiblePanel}>
        <summary className={styles.collapsibleSummary}>タグ一覧から直接選ぶ</summary>
        <div className={styles.filterTagList}>
          {groupedTags.map(([groupName, groupTags]) => (
            <div key={groupName} className={styles.filterTagGroup}>
              <div className={styles.filterTagGroupTitle}>{groupName}</div>
              <div className={styles.filterTagItems}>
                {groupTags.map((tag) => (
                  <label key={tag.id} className={styles.filterTagItem}>
                    <input
                      type="checkbox"
                      checked={filterTagIDs.includes(tag.id)}
                      onChange={() => onToggleFilterTag(tag.id)}
                    />
                    <span>{tag.name}</span>
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>
      </details>

      <div className={styles.filterActions}>
        <button className={styles.secondaryButton} onClick={() => void onSearch()} disabled={loading || snapshotLoading}>
          検索
        </button>
        <button className={styles.successButton} onClick={() => void onSaveAsTarget()} disabled={loading || targetSaving}>
          検索条件を保存
        </button>
        <button className={styles.backButton} onClick={() => void onClear()} disabled={loading}>
          クリア
        </button>
      </div>
    </div>
  );
};

export default ActionsSearchSection;
