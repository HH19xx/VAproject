import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  useWorldSignals,
  type ExternalDataSource,
  type FetchWorldSignalInput,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
import styles from "../../assets/styles/ExternalDataSources.module.scss";

const ExternalDataSources = () => {
  const navigate = useNavigate();
  const { fetchOpenMeteo, loading, error } = useWorldSignals();

  const [source, setSource] = useState<ExternalDataSource>("open_meteo");
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
  const [latitude, setLatitude] = useState("35.6895");
  const [longitude, setLongitude] = useState("139.6917");
  const [pastDays, setPastDays] = useState("14");
  const [forecastDays, setForecastDays] = useState("1");
  const [signalType, setSignalType] = useState("population_total");
  const [pageMessage, setPageMessage] = useState<string | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);

  useEffect(() => {
    setPageMessage(null);
    setPageError(null);

    if (source === "e_stat_dashboard") {
      if (locationKey === "tokyo_shinjuku") {
        setLocationKey("13000");
      }
      setLatitude("0");
      setLongitude("0");
      return;
    }

    if (locationKey === "13000") {
      setLocationKey("tokyo_shinjuku");
    }
    if (latitude === "0") {
      setLatitude("35.6895");
    }
    if (longitude === "0") {
      setLongitude("139.6917");
    }
  }, [latitude, locationKey, longitude, source]);

  const handleFetch = async () => {
    setPageError(null);
    setPageMessage(null);

    const payload: FetchWorldSignalInput = {
      source,
      locationKey: locationKey.trim(),
      latitude: source === "open_meteo" ? Number(latitude) : 0,
      longitude: source === "open_meteo" ? Number(longitude) : 0,
      pastDays: source === "open_meteo" ? Number(pastDays) : 0,
      forecastDays: source === "open_meteo" ? Number(forecastDays) : 0,
    };

    if (!payload.locationKey) {
      setPageError(
        source === "e_stat_dashboard"
          ? "都道府県コードを入力してください。例: 13000"
          : "location_key を入力してください。",
      );
      return;
    }

    if (
      source === "open_meteo" &&
      (!Number.isFinite(payload.latitude) ||
        !Number.isFinite(payload.longitude) ||
        !Number.isFinite(payload.pastDays) ||
        !Number.isFinite(payload.forecastDays))
    ) {
      setPageError("Open-Meteo では緯度・経度・取得日数を数値で入力してください。");
      return;
    }

    const ok = await fetchOpenMeteo(payload);
    if (!ok) {
      setPageError("外部データの取得に失敗しました。サーバログを確認してください。");
      return;
    }

    setPageMessage(
      source === "e_stat_dashboard"
        ? `e-Stat Dashboard のデータを取得しました。${signalType} を行動記録画面と統計分析画面で使えます。`
        : "Open-Meteo のデータを取得しました。行動記録画面と統計分析画面で使えます。",
    );
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>外部データ追加</h1>
          <p className={styles.subtitle}>
            外部データを取得して保存します。保存したデータは行動記録画面と統計分析画面で使えます。
          </p>
        </div>
        <button className={styles.secondaryButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードへ戻る
        </button>
      </div>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>取得設定</h2>
        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>データソース</span>
            <select value={source} onChange={(event) => setSource(event.target.value as ExternalDataSource)}>
              <option value="open_meteo">Open-Meteo</option>
              <option value="e_stat_dashboard">e-Stat Dashboard</option>
            </select>
          </label>

          <label className={styles.field}>
            <span>location_key</span>
            <input value={locationKey} onChange={(event) => setLocationKey(event.target.value)} />
          </label>

          {source === "e_stat_dashboard" && (
            <label className={styles.field}>
              <span>指標</span>
              <select value={signalType} onChange={(event) => setSignalType(event.target.value)}>
                {eStatDashboardSignalOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          )}

          {source === "open_meteo" && (
            <>
              <label className={styles.field}>
                <span>latitude</span>
                <input value={latitude} onChange={(event) => setLatitude(event.target.value)} />
              </label>
              <label className={styles.field}>
                <span>longitude</span>
                <input value={longitude} onChange={(event) => setLongitude(event.target.value)} />
              </label>
              <label className={styles.field}>
                <span>past_days</span>
                <input value={pastDays} onChange={(event) => setPastDays(event.target.value)} />
              </label>
              <label className={styles.field}>
                <span>forecast_days</span>
                <input value={forecastDays} onChange={(event) => setForecastDays(event.target.value)} />
              </label>
            </>
          )}
        </div>

        <div className={styles.notice}>
          {source === "e_stat_dashboard"
            ? "e-Stat Dashboard は年次データです。location_key には都道府県コードを入力してください。例: 13000"
            : "Open-Meteo は時系列データです。緯度・経度と取得日数を指定して取得してください。"}
        </div>

        <div className={styles.actions}>
          <button className={styles.primaryButton} onClick={handleFetch} disabled={loading}>
            外部データを取得して保存
          </button>
          <button className={styles.secondaryButton} onClick={() => navigate("/records/actions")}>
            行動記録画面へ
          </button>
          <button className={styles.secondaryButton} onClick={() => navigate("/statistics")}>
            統計分析へ
          </button>
        </div>
      </section>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>利用できる指標</h2>
        <div className={styles.catalogGrid}>
          <div className={styles.catalogCard}>
            <h3>Open-Meteo</h3>
            <ul>
              <li>気温</li>
              <li>降水量</li>
              <li>風速</li>
              <li>天気コード</li>
            </ul>
          </div>
          <div className={styles.catalogCard}>
            <h3>e-Stat Dashboard</h3>
            <ul>
              {eStatDashboardSignalOptions.map((option) => (
                <li key={option.value}>{option.label}</li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>実行メモ</h2>
        {pageMessage && <div className={styles.successBox}>{pageMessage}</div>}
        {(pageError || error) && <div className={styles.errorBox}>{pageError || error}</div>}
        {!pageMessage && !pageError && !error && (
          <p className={styles.summary}>まだ外部データ取得を実行していません。</p>
        )}
      </section>
    </div>
  );
};

export default ExternalDataSources;
