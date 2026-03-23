export type ExternalSignalOption = {
  value: string;
  label: string;
};

export const eStatDashboardSignalOptions: ExternalSignalOption[] = [
  { value: "population_total", label: "総人口" },
  { value: "youth_ratio", label: "年少人口比（0-14歳）" },
  { value: "senior_ratio", label: "高齢人口比（65歳以上）" },
  { value: "productive_age_ratio", label: "生産年齢人口比" },
];
