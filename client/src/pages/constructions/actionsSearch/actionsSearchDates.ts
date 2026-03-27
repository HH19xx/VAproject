const OPEN_DATA_DAY_MS = 24 * 60 * 60 * 1000;

export const toLocalDateTimeValue = (value: Date): string => {
  const local = new Date(value.getTime() - value.getTimezoneOffset() * 60 * 1000);
  return local.toISOString().slice(0, 16);
};

export const isValidDateRange = (from: string, to: string): boolean => {
  if (!from || !to) return false;
  const fromDate = new Date(from);
  const toDate = new Date(to);
  return Number.isFinite(fromDate.getTime()) && Number.isFinite(toDate.getTime()) && fromDate < toDate;
};

export const syncDateRangeWithOpenDataWindow = (
  pastDaysValue: string,
  forecastDaysValue: string
): { from: string; to: string } | null => {
  const past = Number(pastDaysValue);
  const forecast = Number(forecastDaysValue);
  if (!Number.isFinite(past) || !Number.isFinite(forecast)) return null;

  const now = new Date();
  const from = new Date(now.getTime() - past * OPEN_DATA_DAY_MS);
  const to = new Date(now.getTime() + forecast * OPEN_DATA_DAY_MS);
  return {
    from: toLocalDateTimeValue(from),
    to: toLocalDateTimeValue(to),
  };
};

export const isSameDateTimeByMinute = (left: string, right: string): boolean => {
  const leftTime = new Date(left).getTime();
  const rightTime = new Date(right).getTime();
  if (!Number.isFinite(leftTime) || !Number.isFinite(rightTime)) return false;
  return Math.floor(leftTime / 60000) === Math.floor(rightTime / 60000);
};
