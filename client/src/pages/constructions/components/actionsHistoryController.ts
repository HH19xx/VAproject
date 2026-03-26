import type { AnalysisSeverity, AnalysisSnapshot } from "../../../hooks/useAnalysisSnapshots";
import type { HistorySortKey } from "./actionsHistorySectionHelpers";

export const loadAnalysisHistory = async (
  listSnapshots: (limit?: number) => Promise<AnalysisSnapshot[]>,
  limit = 100
): Promise<AnalysisSnapshot[]> => listSnapshots(limit);

export const buildFilteredSortedHistory = (
  analysisHistory: AnalysisSnapshot[],
  historySeverityFilter: AnalysisSeverity | "ALL",
  historySort: HistorySortKey
): AnalysisSnapshot[] => {
  const filtered =
    historySeverityFilter === "ALL"
      ? analysisHistory
      : analysisHistory.filter((item) => item.severity === historySeverityFilter);

  const copied = [...filtered];
  if (historySort === "created_desc") {
    copied.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
  } else if (historySort === "created_asc") {
    copied.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
  } else {
    copied.sort((a, b) => b.score - a.score);
  }
  return copied;
};

export const clearAnalysisHistory = async (
  clearSnapshots: () => Promise<boolean>
): Promise<boolean> => clearSnapshots();
