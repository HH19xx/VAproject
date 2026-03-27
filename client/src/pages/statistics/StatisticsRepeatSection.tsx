import styles from "../../assets/styles/StatisticsAnalysis.module.scss";
import { formatNumber, repeatBucketLabel } from "./analysis/analysisHelpers";
import type { RepeatBehaviorResult, RepeatBucketUnit } from "./analysis/analysisTypes";

type StatisticsRepeatSectionProps = {
  repeatKeyword: string;
  setRepeatKeyword: (value: string) => void;
  repeatFrom: string;
  setRepeatFrom: (value: string) => void;
  repeatTo: string;
  setRepeatTo: (value: string) => void;
  repeatUnit: RepeatBucketUnit;
  setRepeatUnit: (value: RepeatBucketUnit) => void;
  repeatExpected: string;
  setRepeatExpected: (value: string) => void;
  handleRepeatBehavior: () => void;
  loading: boolean;
  repeatResult: RepeatBehaviorResult | null;
};

const StatisticsRepeatSection = ({
  repeatKeyword,
  setRepeatKeyword,
  repeatFrom,
  setRepeatFrom,
  repeatTo,
  setRepeatTo,
  repeatUnit,
  setRepeatUnit,
  repeatExpected,
  setRepeatExpected,
  handleRepeatBehavior,
  loading,
  repeatResult,
}: StatisticsRepeatSectionProps) => {
  return (
    <section className={styles.panel}>
      <h2 className={styles.sectionTitle}>{"反復行動の推定と検定"}</h2>
      <p className={styles.description}>
        {"同じ行動が繰り返されると想定した場合の、期間あたりの平均期待値を推定します。"}
      </p>
      <div className={styles.formGrid}>
        <label className={styles.field}>
          <span>{"対象キーワード"}</span>
          <input
            value={repeatKeyword}
            onChange={(event) => setRepeatKeyword(event.target.value)}
            placeholder={"例: 通勤 / 運動 / 購買"}
          />
        </label>
        <label className={styles.field}>
          <span>{"開始"}</span>
          <input
            type="datetime-local"
            value={repeatFrom}
            onChange={(event) => setRepeatFrom(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"終了"}</span>
          <input
            type="datetime-local"
            value={repeatTo}
            onChange={(event) => setRepeatTo(event.target.value)}
          />
        </label>
        <label className={styles.field}>
          <span>{"集計単位"}</span>
          <select
            value={repeatUnit}
            onChange={(event) => setRepeatUnit(event.target.value as RepeatBucketUnit)}
          >
            <option value="day">{"日"}</option>
            <option value="week">{"週"}</option>
          </select>
        </label>
        <label className={styles.field}>
          <span>
            {"期待値 / "}
            {repeatBucketLabel(repeatUnit)}
          </span>
          <input
            value={repeatExpected}
            onChange={(event) => setRepeatExpected(event.target.value)}
          />
        </label>
      </div>
      <div className={styles.actions}>
        <button className={styles.primaryButton} onClick={handleRepeatBehavior} disabled={loading}>
          {"反復推定を実行"}
        </button>
      </div>

      {repeatResult && (
        <div className={styles.resultGrid}>
          <div className={styles.resultCard}>
            <h3>{"要約結果"}</h3>
            <p>{"一致した記録数"}: {repeatResult.matchedActionCount}</p>
            <p>{"バケット数"}: {repeatResult.bucketCount}</p>
            <p>{"平均値"}: {formatNumber(repeatResult.observedMean)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"推定区間"}</h3>
            <p>{"標準偏差"}: {formatNumber(repeatResult.stdDev)}</p>
            <p>{"95%下限"}: {formatNumber(repeatResult.confidenceLow)}</p>
            <p>{"95%上限"}: {formatNumber(repeatResult.confidenceHigh)}</p>
          </div>
          <div className={styles.resultCard}>
            <h3>{"検定"}</h3>
            <p>{"期待値"}: {formatNumber(repeatResult.expectedMean)}</p>
            <p>
              {"t値"}:{" "}
              {repeatResult.tStatistic === null ? "-" : formatNumber(repeatResult.tStatistic)}
            </p>
            <p>
              {"両側p値"}:{" "}
              {repeatResult.pValueApprox === null
                ? "-"
                : formatNumber(repeatResult.pValueApprox, 4)}
            </p>
          </div>
        </div>
      )}
    </section>
  );
};

export default StatisticsRepeatSection;
