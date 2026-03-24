import type { ReactElement } from "react";
import type { DistributionBin } from "../../../../hooks/useDistributionAnalysis";

const SVG_VIEWBOX_WIDTH = 900;
const SVG_VIEWBOX_HEIGHT = 260;
const SVG_MARGIN_LEFT = 40;
const SVG_MARGIN_TOP = 30;
const SVG_DRAW_WIDTH = 820;
const SVG_DRAW_HEIGHT = 180;
const GAP_ZERO_LINE_Y = 130;
const GAP_BAR_HALF_WIDTH = 12;

const formatNum = (value: number, digits = 3) => value.toFixed(digits);

export const renderDistributionOverviewSvg = (
  bins: DistributionBin[],
  selectedIndex: number | null,
  onSelect: (index: number) => void
): ReactElement => {
  const maxObserved = Math.max(1, ...bins.map((bin) => bin.observed_count), ...bins.map((bin) => bin.expected_count));
  const binWidth = SVG_DRAW_WIDTH / Math.max(1, bins.length);

  return (
    <svg className="scatterSvg" viewBox={`0 0 ${SVG_VIEWBOX_WIDTH} ${SVG_VIEWBOX_HEIGHT}`} preserveAspectRatio="none">
      {bins.map((bin, index) => {
        const barHeight = (bin.observed_count / maxObserved) * SVG_DRAW_HEIGHT;
        const x = SVG_MARGIN_LEFT + index * binWidth;
        const y = SVG_MARGIN_TOP + (SVG_DRAW_HEIGHT - barHeight);
        const isActive = selectedIndex === index;
        return (
          <rect
            key={`dist-bin-${bin.start}-${bin.end}`}
            x={x + 2}
            y={y}
            width={Math.max(2, binWidth - 4)}
            height={barHeight}
            fill={isActive ? "#0056b3" : "#66b2ff"}
            opacity="0.85"
            role="button"
            tabIndex={0}
            onClick={() => onSelect(index)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onSelect(index);
              }
            }}
          >
            <title>{`${formatNum(bin.start, 2)} - ${formatNum(bin.end, 2)} / 観測=${formatNum(bin.observed_count, 2)} / 期待=${formatNum(bin.expected_count, 2)}`}</title>
          </rect>
        );
      })}
      <polyline
        fill="none"
        stroke="#ff7f0e"
        strokeWidth="3"
        points={bins
          .map((bin, index) => {
            const x = SVG_MARGIN_LEFT + index * binWidth + binWidth / 2;
            const y = SVG_MARGIN_TOP + (SVG_DRAW_HEIGHT - (bin.expected_count / maxObserved) * SVG_DRAW_HEIGHT);
            return `${x},${y}`;
          })
          .join(" ")}
      />
    </svg>
  );
};

export const renderGapSvg = (
  bins: DistributionBin[],
  selectedIndex: number | null,
  onSelect: (index: number) => void
): ReactElement => {
  const maxGap = Math.max(1, ...bins.map((bin) => Math.abs(bin.gap_count)));
  const binWidth = SVG_DRAW_WIDTH / Math.max(1, bins.length);

  return (
    <svg className="scatterSvg" viewBox={`0 0 ${SVG_VIEWBOX_WIDTH} ${SVG_VIEWBOX_HEIGHT}`} preserveAspectRatio="none">
      <line
        x1={SVG_MARGIN_LEFT}
        y1={GAP_ZERO_LINE_Y}
        x2={SVG_MARGIN_LEFT + SVG_DRAW_WIDTH}
        y2={GAP_ZERO_LINE_Y}
        stroke="#666"
        strokeDasharray="4 3"
      />
      {bins.map((bin, index) => {
        const x = SVG_MARGIN_LEFT + index * binWidth + binWidth / 2;
        const height = (Math.abs(bin.gap_count) / maxGap) * 90;
        const shortage = bin.gap_count > 0;
        const y = shortage ? GAP_ZERO_LINE_Y - height : GAP_ZERO_LINE_Y;
        const isActive = selectedIndex === index;
        return (
          <rect
            key={`gap-bin-${bin.start}-${bin.end}`}
            x={x - GAP_BAR_HALF_WIDTH}
            y={y}
            width={GAP_BAR_HALF_WIDTH * 2}
            height={height}
            fill={shortage ? (isActive ? "#198754" : "#7bc67b") : isActive ? "#b02a37" : "#ff8a80"}
            role="button"
            tabIndex={0}
            onClick={() => onSelect(index)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onSelect(index);
              }
            }}
          >
            <title>{`${formatNum(bin.start, 2)} - ${formatNum(bin.end, 2)} / 差分=${formatNum(bin.gap_count, 2)}`}</title>
          </rect>
        );
      })}
    </svg>
  );
};
