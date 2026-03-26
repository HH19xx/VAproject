import { useNavigate } from "react-router-dom";
import styles from "../../assets/styles/Actions.module.scss";
import type { ActionLog } from "../../hooks/useActions";
import { toActionAxisLabel } from "./components/actionsDistribution/actionsDistributionLabels";

interface ActionsResultsSectionProps {
  currentPage: number;
  pageSize: number;
  paginationTotal: number;
  actions: ActionLog[];
  actionAxisCandidates: string[];
  actionScatterXAxis: string;
  actionScatterYAxis: string;
  actionScatterPoints: Array<{ id: number; title: string; x: number; y: number }>;
  prototypeNameMap: Map<number, string>;
  formatTagNames: (tagIDs: number[]) => string;
  loading: boolean;
  onActionScatterXAxisChange: (value: string) => void;
  onActionScatterYAxisChange: (value: string) => void;
  onPageChange: (page: number) => void | Promise<void>;
  onDelete: (id: number) => void | Promise<void>;
}

const ActionsResultsSection = ({
  currentPage,
  pageSize,
  paginationTotal,
  actions,
  actionAxisCandidates,
  actionScatterXAxis,
  actionScatterYAxis,
  actionScatterPoints,
  prototypeNameMap,
  formatTagNames,
  loading,
  onActionScatterXAxisChange,
  onActionScatterYAxisChange,
  onPageChange,
  onDelete,
}: ActionsResultsSectionProps) => {
  const navigate = useNavigate();
  const totalPages = Math.max(1, Math.ceil(paginationTotal / Math.max(pageSize, 1)));
  const canGoPrev = currentPage > 1;
  const canGoNext = currentPage < totalPages;

  return (
    <section className={styles.resultsPanel}>
      <div className={styles.resultsHeader}>
        <div className={styles.analysisTitle}>検索結果の把握</div>
        <div className={styles.analysisRow}>
          検索結果: 全件 {paginationTotal} 件 / 現在ページ {currentPage} / 表示中 {actions.length} 件
        </div>
      </div>

      <div className={styles.resultsGrid}>
        <div className={styles.resultsPrimary}>
          <div className={styles.resultsListPanel}>
            <div className={styles.resultsListHeader}>
              <div className={styles.analysisTitle}>生の検索結果</div>
              <div className={styles.paginationBar}>
                <button
                  type="button"
                  className={styles.suggestionChip}
                  onClick={() => void onPageChange(currentPage - 1)}
                  disabled={!canGoPrev || loading}
                >
                  前へ
                </button>
                <div className={styles.paginationLabel}>
                  {currentPage} / {totalPages}
                </div>
                <button
                  type="button"
                  className={styles.suggestionChip}
                  onClick={() => void onPageChange(currentPage + 1)}
                  disabled={!canGoNext || loading}
                >
                  次へ
                </button>
              </div>
            </div>

            {actions.length === 0 ? (
              <div className={styles.empty}>行動記録がありません。</div>
            ) : (
              <div className={styles.timeline}>
                {actions.map((action, index) => (
                  <div key={action.id} className={styles.timelineItem}>
                    <div className={styles.timelineIndex}>{(currentPage - 1) * pageSize + index + 1}</div>
                    <div className={styles.timelineBody}>
                      <div className={styles.timelineHeader}>
                        <div className={styles.timelineTitle}>{action.title}</div>
                        <div className={styles.timelineTime}>{new Date(action.occurred_at).toLocaleString("ja-JP")}</div>
                      </div>
                      <div className={styles.timelineMeta}>タグ: {formatTagNames(action.tag_ids)}</div>
                      <div className={styles.timelineMeta}>
                        プロトタイプ: {action.prototype_id ? prototypeNameMap.get(action.prototype_id) || `#${action.prototype_id}` : "-"}
                      </div>
                      {action.notes && <div className={styles.timelineNote}>{action.notes}</div>}
                      <div className={styles.timelineActions}>
                        <button className={styles.successButton} onClick={() => navigate(`/records/actions/${action.id}/edit`)}>
                          編集
                        </button>
                        <button className={styles.dangerButton} onClick={() => void onDelete(action.id)} disabled={loading}>
                          削除
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className={styles.resultsSecondary}>
          <div className={styles.vizPanel}>
            <div className={styles.vizHeader}>
              <div className={styles.filterTitle}>検索結果の 2D 可視化</div>
            </div>
            <div className={styles.analysisMetaCard}>
              <div className={styles.analysisMetaLabel}>概要</div>
              <div className={styles.analysisMetaValue}>
                可視化で全体像を先に確認し、必要な場合だけ左の生データをページ単位で見ます。
              </div>
            </div>
            <div className={styles.controlGrid}>
              <label>
                X軸
                <select className={styles.select} value={actionScatterXAxis} onChange={(event) => onActionScatterXAxisChange(event.target.value)}>
                  {actionAxisCandidates.map((axis) => (
                    <option key={`action-x-${axis}`} value={axis}>
                      {toActionAxisLabel(axis)}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Y軸
                <select className={styles.select} value={actionScatterYAxis} onChange={(event) => onActionScatterYAxisChange(event.target.value)}>
                  {actionAxisCandidates.map((axis) => (
                    <option key={`action-y-${axis}`} value={axis}>
                      {toActionAxisLabel(axis)}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <div className={styles.analysisRow}>描画点数: {actionScatterPoints.length} 件</div>
            <div className={styles.analysisRow}>軸キー: x={actionScatterXAxis} / y={actionScatterYAxis}</div>
            {actionScatterPoints.length === 0 && (
              <div className={styles.info}>数値軸として使える値がまだ無いため、散布図を描画できません。</div>
            )}
            <div className={styles.scatterWrapper}>
              <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
                {actionScatterPoints.length > 0 &&
                  (() => {
                    const minX = Math.min(...actionScatterPoints.map((point) => point.x));
                    const maxX = Math.max(...actionScatterPoints.map((point) => point.x));
                    const minY = Math.min(...actionScatterPoints.map((point) => point.y));
                    const maxY = Math.max(...actionScatterPoints.map((point) => point.y));
                    const safeDx = maxX - minX || 1;
                    const safeDy = maxY - minY || 1;
                    return actionScatterPoints.map((point) => {
                      const x = 40 + ((point.x - minX) / safeDx) * 820;
                      const y = 220 - ((point.y - minY) / safeDy) * 180;
                      return (
                        <circle key={`action-point-${point.id}`} className={styles.scatterPoint} cx={x} cy={y} r={4}>
                          <title>{`${point.title} / x=${point.x.toFixed(2)} / y=${point.y.toFixed(2)}`}</title>
                        </circle>
                      );
                    });
                  })()}
                <text x="40" y="20" className={styles.axisLabel}>
                  {`散布図 / x:${toActionAxisLabel(actionScatterXAxis)} / y:${toActionAxisLabel(actionScatterYAxis)}`}
                </text>
              </svg>
            </div>
          </div>
        </div>
      </div>

      <div className={styles.footer}>
        <button className={styles.backButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードへ戻る
        </button>
      </div>
    </section>
  );
};

export default ActionsResultsSection;
