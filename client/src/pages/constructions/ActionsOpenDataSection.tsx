import styles from "../../assets/styles/Actions.module.scss";
import type { AnalysisContextResult, ExternalDataSource } from "../../hooks/useWorldSignals";
import type { DistributionDataset } from "../../hooks/useDistributionAnalysis";
import type { ContextVizMode, ScatterAxisKey } from "./components/actionsOpenDataSectionHelpers";
import { axisLabelMap } from "./components/actionsOpenDataSectionHelpers";

interface ActionsOpenDataSectionProps {
  open: boolean;
  externalDataSource: ExternalDataSource;
  externalSignalType: string;
  effectiveExternalSignalType: string;
  locationKey: string;
  latitude: string;
  longitude: string;
  pastDays: string;
  forecastDays: string;
  worldSignalLoading: boolean;
  worldSignalError: string | null;
  distributionError: string | null;
  distributionDataset: DistributionDataset;
  contextResult: AnalysisContextResult | null;
  contextIsStale: boolean;
  contextVizMode: ContextVizMode;
  scatterXAxis: ScatterAxisKey;
  scatterYAxis: ScatterAxisKey;
  availableWorldAxisEntries: Array<[ScatterAxisKey, string]>;
  contextTimePoints: Array<{ t: number; value: number }>;
  contextScatterPoints: Array<{ x: number; y: number }>;
  onToggleOpen: (open: boolean) => void;
  onExternalDataSourceChange: (value: ExternalDataSource) => void;
  onExternalSignalTypeChange: (value: string) => void;
  onLocationKeyChange: (value: string) => void;
  onLatitudeChange: (value: string) => void;
  onLongitudeChange: (value: string) => void;
  onPastDaysChange: (value: string) => void;
  onForecastDaysChange: (value: string) => void;
  onRefreshOpenDataView: () => void | Promise<void>;
  onLoadAnalysisContext: () => void | Promise<void>;
  onSwitchDistributionDataset: (dataset: DistributionDataset) => void;
  onContextVizModeChange: (mode: ContextVizMode) => void;
  onScatterXAxisChange: (axis: ScatterAxisKey) => void;
  onScatterYAxisChange: (axis: ScatterAxisKey) => void;
  signalOptions: Array<{ value: string; label: string }>;
}

const ActionsOpenDataSection = ({
  open,
  externalDataSource,
  externalSignalType,
  effectiveExternalSignalType,
  locationKey,
  latitude,
  longitude,
  pastDays,
  forecastDays,
  worldSignalLoading,
  worldSignalError,
  distributionError,
  distributionDataset,
  contextResult,
  contextIsStale,
  contextVizMode,
  scatterXAxis,
  scatterYAxis,
  availableWorldAxisEntries,
  contextTimePoints,
  contextScatterPoints,
  onToggleOpen,
  onExternalDataSourceChange,
  onExternalSignalTypeChange,
  onLocationKeyChange,
  onLatitudeChange,
  onLongitudeChange,
  onPastDaysChange,
  onForecastDaysChange,
  onRefreshOpenDataView,
  onLoadAnalysisContext,
  onSwitchDistributionDataset,
  onContextVizModeChange,
  onScatterXAxisChange,
  onScatterYAxisChange,
  signalOptions,
}: ActionsOpenDataSectionProps) => {
  return (
    <details
      className={styles.collapsiblePanel}
      open={open}
      onToggle={(event) => onToggleOpen((event.currentTarget as HTMLDetailsElement).open)}
    >
      <summary className={styles.collapsibleSummary}>
        <span>外部データ連携</span>
        <span className={styles.collapsibleSummaryMeta}>
          {externalDataSource === "open_meteo" ? "Open-Meteo" : "e-Stat Dashboard"}
          {contextResult ? ` / 保存件数 ${contextResult.summary.signal_count}` : ""}
        </span>
      </summary>

      <div className={styles.collapsibleHint}>
        必要なときだけ開いて利用してください。通常は行動記録の検索と分析が主導線です。
      </div>

      {worldSignalError && <div className={styles.errorBox}>外部データエラー: {worldSignalError}</div>}
      {distributionDataset === "world_signals" && distributionError && (
        <div className={styles.errorBox}>外部データ分析エラー: {distributionError}</div>
      )}

      <div className={styles.analysisPanel}>
        <div className={styles.analysisTitle}>取得設定</div>
        <div className={styles.controlGrid}>
          <label>
            data_source
            <select
              className={styles.select}
              value={externalDataSource}
              onChange={(event) => onExternalDataSourceChange(event.target.value as ExternalDataSource)}
            >
              <option value="open_meteo">Open-Meteo</option>
              <option value="e_stat_dashboard">e-Stat Dashboard</option>
            </select>
          </label>
          <label>
            location_key
            <input className={styles.input} value={locationKey} onChange={(event) => onLocationKeyChange(event.target.value)} />
          </label>
          {externalDataSource === "e_stat_dashboard" && (
            <label>
              signal_type
              <select
                className={styles.select}
                value={externalSignalType}
                onChange={(event) => onExternalSignalTypeChange(event.target.value)}
              >
                {signalOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          )}
          <label>
            latitude
            <input
              className={styles.input}
              value={latitude}
              onChange={(event) => onLatitudeChange(event.target.value)}
              disabled={externalDataSource !== "open_meteo"}
            />
          </label>
          <label>
            longitude
            <input
              className={styles.input}
              value={longitude}
              onChange={(event) => onLongitudeChange(event.target.value)}
              disabled={externalDataSource !== "open_meteo"}
            />
          </label>
          <label>
            past_days
            <input
              className={styles.input}
              value={pastDays}
              onChange={(event) => onPastDaysChange(event.target.value)}
              disabled={externalDataSource !== "open_meteo"}
            />
          </label>
          <label>
            forecast_days
            <input
              className={styles.input}
              value={forecastDays}
              onChange={(event) => onForecastDaysChange(event.target.value)}
              disabled={externalDataSource !== "open_meteo"}
            />
          </label>
        </div>

        <div className={styles.info}>
          {externalDataSource === "open_meteo"
            ? "`past_days / forecast_days` を変更すると、表示範囲の `from / to` も自動で同期されます。"
            : "e-Stat Dashboard では `location_key` に都道府県コードを指定してください。例: 13000"}
        </div>

        <div className={styles.filterActions}>
          <button className={styles.successButton} onClick={() => void onRefreshOpenDataView()} disabled={worldSignalLoading}>
            外部データ取得して表示更新
          </button>
          <button className={styles.secondaryButton} onClick={() => void onLoadAnalysisContext()} disabled={worldSignalLoading}>
            分析コンテキスト取得
          </button>
        </div>

        <div className={styles.info}>
          通常は `外部データ取得して表示更新` を使ってください。`分析コンテキスト取得` は保存済みデータの再読込用です。
        </div>
      </div>

      {contextResult && (
        <>
          <div className={styles.analysisPanel}>
            <div className={styles.analysisTitle}>表示中データ</div>
            {contextIsStale && (
              <div className={styles.errorBox}>
                表示中のコンテキストは現在の入力条件と一致していません。`外部データ取得して表示更新` または `分析コンテキスト取得`
                を実行して更新してください。
              </div>
            )}
            <div className={styles.analysisRow}>
              表示期間: {new Date(contextResult.from).toLocaleString("ja-JP")} - {new Date(contextResult.to).toLocaleString("ja-JP")}
            </div>
            <div className={styles.analysisRow}>action_count: {contextResult.summary.action_count}</div>
            <div className={styles.analysisRow}>signal_count: {contextResult.summary.signal_count}</div>
            <div className={styles.analysisRow}>
              source: {String(contextResult.meta?.source || externalDataSource)} / signal_type:{" "}
              {String(contextResult.meta?.signal_type || effectiveExternalSignalType || "-")}
            </div>
            <div className={styles.analysisRow}>
              avg_tags_per_action: {contextResult.summary.avg_tags_per_action.toFixed(2)}
            </div>
            <div className={styles.analysisRow}>
              avg_temperature_c:{" "}
              {typeof contextResult.summary.avg_temperature_c === "number"
                ? contextResult.summary.avg_temperature_c.toFixed(2)
                : "-"}
            </div>
            <div className={styles.analysisRow}>
              total_precipitation_mm: {contextResult.summary.total_precipitation_mm.toFixed(2)}
            </div>
            <div className={styles.analysisRow}>
              avg_signal_value:{" "}
              {typeof contextResult.summary.avg_signal_value === "number"
                ? contextResult.summary.avg_signal_value.toFixed(2)
                : "-"}
            </div>
            <div className={styles.filterActions}>
              <button
                className={styles.secondaryButton}
                onClick={() => onSwitchDistributionDataset("world_signals")}
                disabled={contextResult.summary.signal_count === 0}
              >
                この結果で外部ビッグデータ分析へ切替
              </button>
              <button
                className={styles.backButton}
                onClick={() => onSwitchDistributionDataset("action_logs")}
                disabled={contextResult.summary.action_count === 0}
              >
                行動記録分析へ切替
              </button>
            </div>
          </div>

          <div className={styles.vizModeSwitch}>
            <button
              className={contextVizMode === "time" ? styles.modeActive : styles.modeButton}
              onClick={() => onContextVizModeChange("time")}
            >
              1D 時系列
            </button>
            <button
              className={contextVizMode === "scatter" ? styles.modeActive : styles.modeButton}
              onClick={() => onContextVizModeChange("scatter")}
            >
              2D 散布図
            </button>
          </div>

          {contextVizMode === "time" ? (
            <div className={styles.scatterWrapper}>
              <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
                {contextTimePoints.length > 1 &&
                  (() => {
                    const minT = Math.min(...contextTimePoints.map((point) => point.t));
                    const maxT = Math.max(...contextTimePoints.map((point) => point.t));
                    const minY = Math.min(...contextTimePoints.map((point) => point.value));
                    const maxY = Math.max(...contextTimePoints.map((point) => point.value));
                    const safeDx = maxT - minT || 1;
                    const safeDy = maxY - minY || 1;
                    return contextTimePoints.map((point, index, points) => {
                      if (index === 0) return null;
                      const prev = points[index - 1];
                      const x1 = 40 + ((prev.t - minT) / safeDx) * 820;
                      const y1 = 220 - ((prev.value - minY) / safeDy) * 180;
                      const x2 = 40 + ((point.t - minT) / safeDx) * 820;
                      const y2 = 220 - ((point.value - minY) / safeDy) * 180;
                      return (
                        <line
                          key={`${index}-${point.t}`}
                          x1={x1}
                          y1={y1}
                          x2={x2}
                          y2={y2}
                          stroke="#007bff"
                          strokeWidth="2"
                        />
                      );
                    });
                  })()}
                <text x="40" y="20" className={styles.axisLabel}>
                  {`時系列 / ${new Date(contextResult.from).toLocaleDateString("ja-JP")} - ${new Date(contextResult.to).toLocaleDateString("ja-JP")}`}
                </text>
              </svg>
            </div>
          ) : (
            <>
              <div className={styles.controlGrid}>
                <label>
                  X軸
                  <select
                    className={styles.select}
                    value={scatterXAxis}
                    onChange={(event) => onScatterXAxisChange(event.target.value as ScatterAxisKey)}
                  >
                    {availableWorldAxisEntries.map(([axisKey, axisLabel]) => (
                      <option key={`x-${axisKey}`} value={axisKey}>
                        {axisLabel}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Y軸
                  <select
                    className={styles.select}
                    value={scatterYAxis}
                    onChange={(event) => onScatterYAxisChange(event.target.value as ScatterAxisKey)}
                  >
                    {availableWorldAxisEntries.map(([axisKey, axisLabel]) => (
                      <option key={`y-${axisKey}`} value={axisKey}>
                        {axisLabel}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className={styles.scatterWrapper}>
                <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
                  {contextScatterPoints.length > 0 &&
                    (() => {
                      const minX = Math.min(...contextScatterPoints.map((point) => point.x));
                      const maxX = Math.max(...contextScatterPoints.map((point) => point.x));
                      const minY = Math.min(...contextScatterPoints.map((point) => point.y));
                      const maxY = Math.max(...contextScatterPoints.map((point) => point.y));
                      const safeDx = maxX - minX || 1;
                      const safeDy = maxY - minY || 1;
                      return contextScatterPoints.map((point, index) => {
                        const x = 40 + ((point.x - minX) / safeDx) * 820;
                        const y = 220 - ((point.y - minY) / safeDy) * 180;
                        return <circle key={`context-point-${index}`} cx={x} cy={y} r={4} className={styles.scatterPoint} />;
                      });
                    })()}
                  <text x="40" y="20" className={styles.axisLabel}>
                    {`散布図 / x:${axisLabelMap[scatterXAxis]} / y:${axisLabelMap[scatterYAxis]}`}
                  </text>
                </svg>
              </div>
            </>
          )}
        </>
      )}
    </details>
  );
};

export default ActionsOpenDataSection;