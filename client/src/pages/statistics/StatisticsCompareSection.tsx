import type { ExternalDataSource } from "../../hooks/useWorldSignals";
import styles from "../../assets/styles/StatisticsAnalysis.module.scss";
import { actionAxisLabel, formatNumber } from "./analysis/analysisHelpers";
import type {
  ActionAxisKey,
  TwoGroupAnalysisResult,
  WorldAxisKey,
} from "./analysis/analysisTypes";

type StatisticsCompareSectionProps = {
  dataset: "action_logs" | "world_signals";
  compareFromA: string;
  setCompareFromA: (value: string) => void;
  compareToA: string;
  setCompareToA: (value: string) => void;
  compareFromB: string;
  setCompareFromB: (value: string) => void;
  compareToB: string;
  setCompareToB: (value: string) => void;
  compareActionAxis: ActionAxisKey;
  setCompareActionAxis: (value: ActionAxisKey) => void;
  compareWorldAxis: WorldAxisKey;
  setCompareWorldAxis: (value: WorldAxisKey) => void;
  actionAxisOptions: ActionAxisKey[];
  source: ExternalDataSource;
  selectedSignalLabel: string;
  loading: boolean;
  handleCompareMeans: () => void;
  compareResult: TwoGroupAnalysisResult | null;
};

const StatisticsCompareSection = ({
  dataset,
  compareFromA,
  setCompareFromA,
  compareToA,
  setCompareToA,
  compareFromB,
  setCompareFromB,
  compareToB,
  setCompareToB,
  compareActionAxis,
  setCompareActionAxis,
  compareWorldAxis,
  setCompareWorldAxis,
  actionAxisOptions,
  source,
  selectedSignalLabel,
  loading,
  handleCompareMeans,
  compareResult,
}: StatisticsCompareSectionProps) => {
  return (
    <section className={styles.panel}>
      <h2 className={styles.sectionTitle}>{"2期間の平均値比較"}</h2>
      <div className={styles.formGrid}>
        <label className={styles.field}>
          <span>{"期間A 開始"}</span>
          <input
            type="datetime-local"
            value={compareFromA}
            onChange={(event) => setCompareFromA(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"期間A 終了"}</span>
          <input
            type="datetime-local"
            value={compareToA}
            onChange={(event) => setCompareToA(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"期間B 開始"}</span>
          <input
            type="datetime-local"
            value={compareFromB}
            onChange={(event) => setCompareFromB(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"期間B 終了"}</span>
          <input
            type="datetime-local"
            value={compareToB}
            onChange={(event) => setCompareToB(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"比較軸"}</span>
          {dataset === "action_logs" ? (
            <select
              value={compareActionAxis}
              onChange={(event) => setCompareActionAxis(event.target.value as ActionAxisKey)}
            >
              {actionAxisOptions.map((axis) => (
                <option key={axis} value={axis}>
                  {actionAxisLabel(axis)}
                </option>
              ))}
            </select>
          ) : (
            <select
              value={compareWorldAxis}
              onChange={(event) => setCompareWorldAxis(event.target.value as WorldAxisKey)}
            >
              {source === "e_stat_dashboard" ? (
                <option value="signal_value">{selectedSignalLabel}</option>
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
        <button className={styles.primaryButton} onClick={handleCompareMeans} disabled={loading}>
          {"平均値比較を実行"}
        </button>
      </div>

      {compareResult && (
        <div className={styles.resultGrid}>
          <div className={styles.resultCard}>
            <h3>{"期間A"}</h3>
            <p>{"件数"}: {compareResult.groupA.count}</p>
            <p>{"平均"}: {formatNumber(compareResult.groupA.mean)}</p>
            <p>{"標準偏差"}: {formatNumber(compareResult.groupA.stdDev)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"期間B"}</h3>
            <p>{"件数"}: {compareResult.groupB.count}</p>
            <p>{"平均"}: {formatNumber(compareResult.groupB.mean)}</p>
            <p>{"標準偏差"}: {formatNumber(compareResult.groupB.stdDev)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"比較結果"}</h3>
            <p>{"平均差"}: {formatNumber(compareResult.meanDiff)}</p>
            <p>{"t値"}: {formatNumber(compareResult.tStatistic)}</p>
            <p>{"自由度"}: {formatNumber(compareResult.degreesOfFreedom, 1)}</p>
            <p>{"両側p値"}: {formatNumber(compareResult.pValueApprox, 4)}</p>
          </div>
        </div>
      )}
    </section>
  );
};

export default StatisticsCompareSection;
