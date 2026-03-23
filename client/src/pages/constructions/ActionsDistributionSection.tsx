import styles from "../../assets/styles/Actions.module.scss";
import type { DistributionAnalysisResult, DistributionDataset } from "../../hooks/useDistributionAnalysis";
import type { ScatterAxisKey } from "./components/actionsOpenDataSectionHelpers";
import {
  axisLabelMap,
  formatNum,
  renderDistributionOverviewSvg,
  renderGapSvg,
  toActionAxisLabel,
  toBadgeClass,
  toDistributionSeverity,
  toIssueClass,
  toResidualAxisReasonLabel,
  toResidualAxisTypeLabel,
  toResidualStrategyLabel,
  type AnalysisViewState,
  type SelectedDistributionBin,
} from "./components/actionsDistributionSectionHelpers";

interface ActionsDistributionSectionProps {
  distributionResult: DistributionAnalysisResult | null;
  distributionDataset: DistributionDataset;
  distributionAxis: string;
  distributionDatasetLabel: string;
  distributionEmptyHint: string | null;
  selectedDatasetCount: number;
  distributionLoading: boolean;
  selectedDistributionBin: SelectedDistributionBin | null;
  selectedBinStatus:
    | {
        kind: "shortage" | "excess" | "balanced";
        label: string;
        message: string;
      }
    | null;
  actionAxisCandidates: string[];
  availableWorldAxisEntries: Array<[ScatterAxisKey, string]>;
  setDistributionAxis: (value: string) => void;
  switchDistributionDataset: (dataset: DistributionDataset) => void;
  runDistributionAnalysisFromCurrent: () => Promise<void>;
  runDistributionFromCurrentInputs: (datasetOverride?: DistributionDataset, axisOverride?: string) => Promise<void>;
  appendSuggestedTagToQuery: (tag: string, mode?: "append" | "search") => Promise<void>;
  selectDistributionBin: (index: number) => void;
  analysisState: AnalysisViewState | null;
}

const ActionsDistributionSection = ({
  distributionResult,
  distributionDataset,
  distributionAxis,
  distributionDatasetLabel,
  distributionEmptyHint,
  selectedDatasetCount,
  distributionLoading,
  selectedDistributionBin,
  selectedBinStatus,
  actionAxisCandidates,
  availableWorldAxisEntries,
  setDistributionAxis,
  switchDistributionDataset,
  runDistributionAnalysisFromCurrent,
  runDistributionFromCurrentInputs,
  appendSuggestedTagToQuery,
  selectDistributionBin,
  analysisState,
}: ActionsDistributionSectionProps) => {
  return (
    <>
      <div className={styles.analysisPanel}>
        <div className={styles.analysisTitle}>
          分布分析（正規分布基準）
          {distributionResult && (
            <span className={toBadgeClass(toDistributionSeverity(distributionResult))}>
              {toDistributionSeverity(distributionResult)}
            </span>
          )}
        </div>
        <div className={styles.analysisRow}>
          現在の分析対象: {distributionDatasetLabel} / 利用可能件数: {selectedDatasetCount}
        </div>
        {distributionEmptyHint && <div className={styles.errorBox}>{distributionEmptyHint}</div>}

        <div className={styles.controlGrid}>
          <label>
            データ集
            <select
              className={styles.select}
              value={distributionDataset}
              onChange={(event) => switchDistributionDataset(event.target.value as DistributionDataset)}
            >
              <option value="action_logs">行動記録</option>
              <option value="world_signals">外部ビッグデータ</option>
            </select>
          </label>
          <label>
            軸
            <select className={styles.select} value={distributionAxis} onChange={(event) => setDistributionAxis(event.target.value)}>
              {distributionDataset === "action_logs"
                ? actionAxisCandidates.map((axis) => (
                    <option key={`distribution-axis-${axis}`} value={axis}>
                      {toActionAxisLabel(axis)}
                    </option>
                  ))
                : availableWorldAxisEntries.map(([axisKey, axisLabel]) => (
                    <option key={`distribution-axis-${axisKey}`} value={axisKey}>
                      {axisLabel}
                    </option>
                  ))}
            </select>
          </label>
        </div>

        <div className={styles.filterActions}>
          <button
            className={styles.successButton}
            onClick={() => switchDistributionDataset(distributionDataset === "action_logs" ? "world_signals" : "action_logs")}
          >
            {distributionDataset === "action_logs" ? "外部ビッグデータへ切替" : "行動記録へ切替"}
          </button>
          <button
            className={styles.secondaryButton}
            onClick={() => void runDistributionAnalysisFromCurrent()}
            disabled={distributionLoading || selectedDatasetCount === 0}
          >
            分布分析を実行
          </button>
        </div>

        {distributionResult && (
          <>
            <div className={styles.analysisRow}>
              対象: {distributionDatasetLabel} / 軸:{" "}
              {distributionResult.dataset === "action_logs"
                ? toActionAxisLabel(distributionResult.axis)
                : axisLabelMap[distributionResult.axis as ScatterAxisKey] || distributionResult.axis}
            </div>

            {distributionResult.meta && (
              <div className={styles.analysisMetaGrid}>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>残差モデル</div>
                  <div className={styles.analysisMetaValue}>{toResidualStrategyLabel(distributionResult.meta.residual_strategy)}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>軸種別</div>
                  <div className={styles.analysisMetaValue}>{toResidualAxisTypeLabel(distributionResult.meta.residual_axis_type)}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>判定理由</div>
                  <div className={styles.analysisMetaValue}>{toResidualAxisReasonLabel(distributionResult.meta.residual_axis_reason)}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>窓幅</div>
                  <div className={styles.analysisMetaValue}>{String(distributionResult.meta.residual_window || "-")}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>サンプル数</div>
                  <div className={styles.analysisMetaValue}>{String(distributionResult.meta.sample_count || "-")}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>ユニーク値数</div>
                  <div className={styles.analysisMetaValue}>{String(distributionResult.meta.unique_value_count || "-")}</div>
                </div>
                <div className={styles.analysisMetaCard}>
                  <div className={styles.analysisMetaLabel}>ユニーク比率</div>
                  <div className={styles.analysisMetaValue}>
                    {typeof distributionResult.meta.unique_ratio === "number"
                      ? formatNum(distributionResult.meta.unique_ratio, 3)
                      : String(distributionResult.meta.unique_ratio || "-")}
                  </div>
                </div>
              </div>
            )}

            {(distributionResult.raw_bins.length > 0 || distributionResult.residual_bins.length > 0) && (
              <>
                <div className={styles.analysisTitle}>0. 補完グラフ</div>
                <div className={styles.analysisRow}>
                  正規分布へ近づくために、どの帯域へどれだけ補完が必要かを先に可視化します。
                </div>

                {selectedDistributionBin && (
                  <div className={styles.selectedBinInfo}>
                    選択帯域: {formatNum(selectedDistributionBin.bin.start, 2)} - {formatNum(selectedDistributionBin.bin.end, 2)} /
                    観測={formatNum(selectedDistributionBin.bin.observed_count, 2)} /
                    期待={formatNum(selectedDistributionBin.bin.expected_count, 2)} /
                    差分={formatNum(selectedDistributionBin.bin.gap_count, 2)}
                  </div>
                )}
                {selectedBinStatus && (
                  <div
                    className={`${styles.selectedBinHint} ${
                      selectedBinStatus.kind === "shortage"
                        ? styles.selectedBinHintShortage
                        : selectedBinStatus.kind === "excess"
                          ? styles.selectedBinHintExcess
                          : styles.selectedBinHintBalanced
                    }`}
                  >
                    {selectedBinStatus.label}: {selectedBinStatus.message}
                  </div>
                )}

                {distributionResult.raw_bins.length > 0 && (
                  <>
                    <div className={styles.analysisTitle}>0-1. 生分布</div>
                    <div className={styles.scatterWrapper}>
                      {renderDistributionOverviewSvg(distributionResult.raw_bins, null, () => {})}
                    </div>
                  </>
                )}

                {distributionResult.residual_bins.length > 0 && (
                  <>
                    <div className={styles.analysisTitle}>0-2. 残差分布</div>
                    <div className={styles.scatterWrapper}>
                      {renderDistributionOverviewSvg(
                        distributionResult.residual_bins,
                        selectedDistributionBin?.index ?? null,
                        selectDistributionBin
                      )}
                    </div>
                    <div className={styles.scatterWrapper}>
                      {renderGapSvg(distributionResult.residual_bins, selectedDistributionBin?.index ?? null, selectDistributionBin)}
                    </div>
                  </>
                )}
              </>
            )}

            <div className={styles.analysisTitle}>1. 逸脱概要</div>
            <div className={styles.analysisRow}>
              正規性スコア={formatNum(distributionResult.residual_current.normality_score, 1)} / 歪度=
              {formatNum(distributionResult.residual_current.skewness, 2)} / 峰数={distributionResult.residual_current.peak_count} /
              件数={distributionResult.residual_current.count}
            </div>
            <div className={styles.analysisRow}>
              生分布の平均={formatNum(distributionResult.current.mean, 2)} / 分散={formatNum(distributionResult.current.variance, 2)} /
              欠損={distributionResult.current.missing_count} / 件数={distributionResult.current.count}
            </div>
            {distributionResult.residual_baseline && (
              <div className={styles.analysisRow}>
                基準比較: 基準正規性スコア={formatNum(distributionResult.residual_baseline.normality_score, 1)} /
                平均差={formatNum(distributionResult.comparison?.mean_diff ?? 0, 2)} /
                分散差={formatNum(distributionResult.comparison?.variance_diff ?? 0, 2)} /
                正規性差={formatNum(distributionResult.comparison?.normality_score_diff ?? 0, 1)}
              </div>
            )}

            <div className={styles.analysisTitle}>2. 隠れた要因</div>
            {distributionResult.hidden_factor_candidates.length === 0 ? (
              <div className={styles.analysisRow}>目立った隠れ要因候補はまだ出ていません。</div>
            ) : (
              distributionResult.hidden_factor_candidates.map((item, index) => (
                <div key={`hidden-factor-${index}`} className={styles.analysisRow}>
                  - {item}
                </div>
              ))
            )}

            <div className={styles.analysisTitle}>3. 追加すべき条件</div>
            {distributionResult.format_suggestions.length === 0 ? (
              <div className={styles.analysisRow}>追加候補はまだありません。</div>
            ) : (
              distributionResult.format_suggestions.map((item, index) => (
                <div key={`format-suggestion-${index}`} className={styles.analysisRow}>
                  - {item}
                </div>
              ))
            )}

            <div className={styles.analysisTitle}>候補タグ</div>
            {distributionResult.suggested_tags.length === 0 ? (
              <div className={styles.analysisRow}>候補タグはありません。</div>
            ) : (
              <div className={styles.suggestionList}>
                {distributionResult.suggested_tags.map((item, index) => (
                  <div
                    key={`suggested-tag-${index}`}
                    className={`${styles.suggestionActionGroup} ${selectedBinStatus ? styles.suggestionActionGroupActive : ""}`}
                  >
                    <button type="button" className={styles.suggestionChip} onClick={() => void appendSuggestedTagToQuery(item, "append")}>
                      + {item}
                    </button>
                    <button type="button" className={styles.suggestionSearchButton} onClick={() => void appendSuggestedTagToQuery(item, "search")}>
                      追加して即検索
                    </button>
                  </div>
                ))}
              </div>
            )}

            <div className={styles.analysisTitle}>候補軸</div>
            {distributionResult.suggested_axes.length === 0 ? (
              <div className={styles.analysisRow}>候補軸はありません。</div>
            ) : (
              <div className={styles.suggestionList}>
                {distributionResult.suggested_axes.map((item, index) => (
                  <button
                    key={`suggested-axis-${index}`}
                    type="button"
                    className={`${styles.suggestionChip} ${selectedBinStatus ? styles.suggestionChipActive : ""}`}
                    onClick={() => void runDistributionFromCurrentInputs(distributionResult.dataset, item)}
                  >
                    {distributionResult.dataset === "action_logs"
                      ? toActionAxisLabel(item)
                      : axisLabelMap[item as ScatterAxisKey] || item}
                  </button>
                ))}
              </div>
            )}

            <div className={styles.analysisTitle}>4. 詳細判定</div>
            {distributionResult.issues.map((issue) => (
              <div key={`${issue.code}-${issue.message}`} className={`${styles.analysisRow} ${toIssueClass(issue.category)}`}>
                [{issue.category}/{issue.severity}] {issue.message}
                {issue.suggestion ? ` / 対応: ${issue.suggestion}` : ""}
              </div>
            ))}
          </>
        )}
      </div>

      {analysisState && (
        <div className={styles.analysisPanel}>
          <div className={styles.analysisTitle}>
            従来分析（補助）
            <span className={toBadgeClass(analysisState.severity)}>{analysisState.severity}</span>
          </div>
          <div className={styles.analysisRow}>
            上段の分布分析を主表示として見てください。以下は旧互換の差分表示です。
          </div>
          <div className={styles.analysisRow}>
            スコア={formatNum(analysisState.score, 2)} / 平均差={formatNum(analysisState.deltaAvgTag, 2)} / 分散差=
            {formatNum(analysisState.deltaVarTag, 2)} / プロトタイプ率差=
            {formatNum(analysisState.deltaPrototypeRate * 100, 2)}%
          </div>
          {analysisState.pValue !== undefined && (
            <div className={styles.analysisRow}>
              p値={analysisState.pValue.toExponential(3)} / 有意={String(analysisState.significant)}
            </div>
          )}
        </div>
      )}
    </>
  );
};

export default ActionsDistributionSection;
