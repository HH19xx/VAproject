import { useState } from "react";
import type { AnalysisContextResult, ExternalDataSource } from "../../../hooks/useWorldSignals";

const useActionsOpenDataState = () => {
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
  const [externalDataSource, setExternalDataSource] = useState<ExternalDataSource>("open_meteo");
  const [externalSignalType, setExternalSignalType] = useState("population_total");
  const [latitude, setLatitude] = useState("35.6895");
  const [longitude, setLongitude] = useState("139.6917");
  const [pastDays, setPastDays] = useState("7");
  const [forecastDays, setForecastDays] = useState("1");
  const [contextResult, setContextResult] = useState<AnalysisContextResult | null>(null);
  const [openDataPanelOpen, setOpenDataPanelOpen] = useState(false);
  const [actionScatterXAxis, setActionScatterXAxis] = useState<string>("occurred_at");
  const [actionScatterYAxis, setActionScatterYAxis] = useState<string>("tag_count");

  return {
    locationKey,
    setLocationKey,
    externalDataSource,
    setExternalDataSource,
    externalSignalType,
    setExternalSignalType,
    latitude,
    setLatitude,
    longitude,
    setLongitude,
    pastDays,
    setPastDays,
    forecastDays,
    setForecastDays,
    contextResult,
    setContextResult,
    openDataPanelOpen,
    setOpenDataPanelOpen,
    actionScatterXAxis,
    setActionScatterXAxis,
    actionScatterYAxis,
    setActionScatterYAxis,
  };
};

export default useActionsOpenDataState;
