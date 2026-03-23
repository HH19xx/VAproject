import styles from "../../assets/styles/Actions.module.scss";
import type { AnalysisSeverity, AnalysisSnapshot } from "../../hooks/useAnalysisSnapshots";
import type { HistorySortKey } from "./components/actionsHistorySectionHelpers";
import { extractSnapshotMeta, formatNum, toDatasetLabel } from "./components/actionsHistorySectionHelpers";

interface ActionsHistorySectionProps {
  analysisHistory: AnalysisSnapshot[];
  filteredSortedHistory: AnalysisSnapshot[];
  historySeverityFilter: AnalysisSeverity | "ALL";
  historySort: HistorySortKey;
  snapshotLoading: boolean;
  onHistorySeverityFilterChange: (value: AnalysisSeverity | "ALL") => void;
  onHistorySortChange: (value: HistorySortKey) => void;
  onClearSnapshots: () => void | Promise<void>;
}

const ActionsHistorySection = ({
  analysisHistory,
  filteredSortedHistory,
  historySeverityFilter,
  historySort,
  snapshotLoading,
  onHistorySeverityFilterChange,
  onHistorySortChange,
  onClearSnapshots,
}: ActionsHistorySectionProps) => {
  return (
    <div className={styles.trendPanel}>
      <div className={styles.analysisTitle}>分析履歴</div>
      <div className={styles.controlGrid}>
        <label>
          重要度フィルタ
          <select
            className={styles.select}
            value={historySeverityFilter}
            onChange={(event) => onHistorySeverityFilterChange(event.target.value as AnalysisSeverity | "ALL")}
          >
            <option value="ALL">ALL</option>
            <option value="OK">OK</option>
            <option value="NOTICE">NOTICE</option>
            <option value="ALERT">ALERT</option>
          </select>
        </label>
        <label>
          並び順
          <select className={styles.select} value={historySort} onChange={(event) => onHistorySortChange(event.target.value as HistorySortKey)}>
            <option value="created_desc">新しい順</option>
            <option value="created_asc">古い順</option>
            <option value="score_desc">スコア順</option>
          </select>
        </label>
      </div>

      <svg className={styles.trendSvg} viewBox="0 0 600 120" preserveAspectRatio="none">
        {filteredSortedHistory.length > 1 &&
          filteredSortedHistory.slice(0, 40).map((entry, index, list) => {
            if (index === 0) return null;
            const maxScore = Math.max(1, ...list.map((item) => item.score));
            const x1 = ((list.length - index) / (list.length - 1)) * 590 + 5;
            const x2 = ((list.length - (index - 1)) / (list.length - 1)) * 590 + 5;
            const y1 = 110 - (entry.score / maxScore) * 100;
            const y2 = 110 - (list[index - 1].score / maxScore) * 100;
            return <line key={entry.id} x1={x1} y1={y1} x2={x2} y2={y2} stroke="#007bff" strokeWidth="2" />;
          })}
      </svg>

      <div className={styles.analysisRow}>履歴件数: {filteredSortedHistory.length} 件 / 全体 {analysisHistory.length} 件</div>
      <div className={styles.analysisRow}>
        直近5件:{" "}
        {filteredSortedHistory
          .slice(0, 5)
          .map((item) => {
            const meta = extractSnapshotMeta(item.query_text);
            return `${new Date(item.created_at).toLocaleString("ja-JP")} [${item.severity}] 逸脱スコア=${formatNum(item.score, 2)} / データ集=${toDatasetLabel(meta.dataset)} / 軸=${meta.axis || "未指定"} / 条件=${meta.query || "なし"}`;
          })
          .join(" / ")}
      </div>

      <div className={styles.filterActions}>
        <button className={styles.dangerButton} onClick={() => void onClearSnapshots()} disabled={snapshotLoading}>
          履歴を全削除
        </button>
      </div>
    </div>
  );
};

export default ActionsHistorySection;