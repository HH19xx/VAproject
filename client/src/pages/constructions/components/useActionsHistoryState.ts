import { useState } from "react";
import type { AnalysisSeverity, AnalysisSnapshot } from "../../../hooks/useAnalysisSnapshots";
import type { HistorySortKey } from "./actionsHistorySectionHelpers";

const useActionsHistoryState = () => {
  const [analysisHistory, setAnalysisHistory] = useState<AnalysisSnapshot[]>([]);
  const [historySeverityFilter, setHistorySeverityFilter] = useState<AnalysisSeverity | "ALL">("ALL");
  const [historySort, setHistorySort] = useState<HistorySortKey>("created_desc");

  return {
    analysisHistory,
    setAnalysisHistory,
    historySeverityFilter,
    setHistorySeverityFilter,
    historySort,
    setHistorySort,
  };
};

export default useActionsHistoryState;
