export const buildContextStaleState = (
  contextResult: {
    location_key: string;
    from: string;
    to: string;
    meta?: { source?: string; signal_type?: string };
  } | null,
  current: {
    source: string;
    signalType: string;
    locationKey: string;
    from: string;
    to: string;
  },
  isSameDateTimeByMinute: (left: string, right: string) => boolean
) => {
  if (!contextResult || !current.from || !current.to) return false;
  const metaSource = String(contextResult.meta?.source || "open_meteo");
  const metaSignalType = String(contextResult.meta?.signal_type || "");
  return (
    contextResult.location_key !== current.locationKey ||
    metaSource !== current.source ||
    metaSignalType !== current.signalType ||
    !isSameDateTimeByMinute(contextResult.from, current.from) ||
    !isSameDateTimeByMinute(contextResult.to, current.to)
  );
};
