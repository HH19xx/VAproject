import type { ExternalDataSource } from "../../hooks/useWorldSignals";
import styles from "../../assets/styles/StatisticsAnalysis.module.scss";
import { actionAxisLabel, formatNumber } from "./analysis/analysisHelpers";
import type {
  ActionAxisKey,
  CorrelationAnalysisResult,
  DatasetKind,
  WorldAxisKey,
} from "./analysis/analysisTypes";

type StatisticsCorrelationSectionProps = {
  dataset: DatasetKind;
  correlationFrom: string;
  setCorrelationFrom: (value: string) => void;
  correlationTo: string;
  setCorrelationTo: (value: string) => void;
  correlationActionXAxis: ActionAxisKey;
  setCorrelationActionXAxis: (value: ActionAxisKey) => void;
  correlationActionYAxis: ActionAxisKey;
  setCorrelationActionYAxis: (value: ActionAxisKey) => void;
  correlationWorldXAxis: WorldAxisKey;
  setCorrelationWorldXAxis: (value: WorldAxisKey) => void;
  correlationWorldYAxis: WorldAxisKey;
  setCorrelationWorldYAxis: (value: WorldAxisKey) => void;
  actionAxisOptions: ActionAxisKey[];
  source: ExternalDataSource;
  selectedSignalLabel: string;
  loading: boolean;
  handleCorrelation: () => void;
  correlationResult: CorrelationAnalysisResult | null;
};

const StatisticsCorrelationSection = ({
  dataset,
  correlationFrom,
  setCorrelationFrom,
  correlationTo,
  setCorrelationTo,
  correlationActionXAxis,
  setCorrelationActionXAxis,
  correlationActionYAxis,
  setCorrelationActionYAxis,
  correlationWorldXAxis,
  setCorrelationWorldXAxis,
  correlationWorldYAxis,
  setCorrelationWorldYAxis,
  actionAxisOptions,
  source,
  selectedSignalLabel,
  loading,
  handleCorrelation,
  correlationResult,
}: StatisticsCorrelationSectionProps) => {
  return (
    <section className={styles.panel}>
      <h2 className={styles.sectionTitle}>{"相関分析"}</h2>
      <div className={styles.formGrid}>
        <label className={styles.field}>
          <span>{"開始"}</span>
          <input
            type="datetime-local"
            value={correlationFrom}
            onChange={(event) => setCorrelationFrom(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"終了"}</span>
          <input
            type="datetime-local"
            value={correlationTo}
            onChange={(event) => setCorrelationTo(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"X軸"}</span>
          {dataset === "action_logs" ? (
            <select
              value={correlationActionXAxis}
              onChange={(event) => setCorrelationActionXAxis(event.target.value as ActionAxisKey)}
            >
              {actionAxisOptions.map((axis) => (
                <option key={axis} value={axis}>
                  {actionAxisLabel(axis)}
                </option>
              ))}
            </select>
          ) : (
            <select
              value={correlationWorldXAxis}
              onChange={(event) => setCorrelationWorldXAxis(event.target.value as WorldAxisKey)}
            >
              {source === "e_stat_dashboard" ? (
                <>
                  <option value="observed_year">{"年"}</option>
                  <option value="signal_value">{selectedSignalLabel}</option>
                </>
              ) : (
                <>
                  <option value="temperature_c">{"気温"}</option>
                  <option value="precipitation_mm">{"降水量"}</option>
                  <option value="wind_speed_ms">{"風速"}</option>
                  <option value="weather_code">{"天気コード"}</option>
                </>
              )}
            </select>
          )}
        </label>
        <label className={styles.field}>
          <span>{"Y軸"}</span>
          {dataset === "action_logs" ? (
            <select
              value={correlationActionYAxis}
              onChange={(event) => setCorrelationActionYAxis(event.target.value as ActionAxisKey)}
            >
              {actionAxisOptions.map((axis) => (
                <option key={axis} value={axis}>
                  {actionAxisLabel(axis)}
                </option>
              ))}
            </select>
          ) : (
            <select
              value={correlationWorldYAxis}
              onChange={(event) => setCorrelationWorldYAxis(event.target.value as WorldAxisKey)}
            >
              {source === "e_stat_dashboard" ? (
                <>
                  <option value="signal_value">{selectedSignalLabel}</option>
                  <option value="observed_year">{"年"}</option>
                </>
              ) : (
                <>
                  <option value="temperature_c">{"気温"}</option>
                  <option value="precipitation_mm">{"降水量"}</option>
                  <option value="wind_speed_ms">{"風速"}</option>
                  <option value="weather_code">{"天気コード"}</option>
                </>
              )}
            </select>
          )}
        </label>
      </div>
      <div className={styles.actions}>
        <button className={styles.primaryButton} onClick={handleCorrelation} disabled={loading}>
          {"相関分析を実行"}
        </button>
      </div>

      {correlationResult && (
        <div className={styles.resultGrid}>
          <div className={styles.resultCard}>
            <h3>{"相関係数"}</h3>
            <p>{formatNumber(correlationResult.correlation, 4)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"共分散"}</h3>
            <p>{formatNumber(correlationResult.covariance, 4)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"概要"}</h3>
            <p>{"対象ペア数"}: {correlationResult.count}</p>
            <p>{"両側p値"}: {formatNumber(correlationResult.pValueApprox, 4)}</p>
          </div>
        </div>
      )}
    </section>
  );
};

export default StatisticsCorrelationSection;
