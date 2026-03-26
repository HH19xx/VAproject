import { useState } from "react";
import type { DistributionAnalysisResult, DistributionDataset } from "../../../hooks/useDistributionAnalysis";
import type { AnalysisViewState } from "./actionsDistribution/actionsDistributionScore";
import type { SelectedDistributionBin } from "./actionsDistribution/actionsDistributionTypes";

const useActionsDistributionState = () => {
  const [analysisState, setAnalysisState] = useState<AnalysisViewState | null>(null);
  const [distributionDataset, setDistributionDataset] = useState<DistributionDataset>("action_logs");
  const [distributionAxis, setDistributionAxis] = useState<string>("tag_count");
  const [distributionResult, setDistributionResult] = useState<DistributionAnalysisResult | null>(null);
  const [selectedDistributionBin, setSelectedDistributionBin] = useState<SelectedDistributionBin | null>(null);
  const [analysisMessage, setAnalysisMessage] = useState<string | null>(null);

  return {
    analysisState,
    setAnalysisState,
    distributionDataset,
    setDistributionDataset,
    distributionAxis,
    setDistributionAxis,
    distributionResult,
    setDistributionResult,
    selectedDistributionBin,
    setSelectedDistributionBin,
    analysisMessage,
    setAnalysisMessage,
  };
};

export default useActionsDistributionState;
