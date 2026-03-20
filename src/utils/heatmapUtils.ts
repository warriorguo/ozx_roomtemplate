import type { Template } from '../types/newTemplate';

const HEATMAP_RADIUS = 5;
const DPS_WEIGHT = 1.0;
const ZONER_WEIGHT = 0.8;

/**
 * Compute a normalized threat heatmap from DPS and Zoner layers.
 * Each DPS/Zoner cell radiates influence to surrounding cells with linear falloff.
 * Returns a 2D array of scores in [0, 1].
 */
export function computeThreatHeatmap(template: Template): number[][] {
  const { width, height, dps, zoner } = template;
  const scores: number[][] = Array.from({ length: height }, () => Array(width).fill(0));

  for (let ey = 0; ey < height; ey++) {
    for (let ex = 0; ex < width; ex++) {
      const isDps = dps[ey][ex] === 1;
      const isZoner = zoner[ey][ex] === 1;
      if (!isDps && !isZoner) continue;

      // A cell can be both DPS and Zoner in theory; take the higher weight
      const weight = isDps ? DPS_WEIGHT : ZONER_WEIGHT;

      for (let dy = -HEATMAP_RADIUS; dy <= HEATMAP_RADIUS; dy++) {
        for (let dx = -HEATMAP_RADIUS; dx <= HEATMAP_RADIUS; dx++) {
          const ny = ey + dy;
          const nx = ex + dx;
          if (ny < 0 || ny >= height || nx < 0 || nx >= width) continue;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist > HEATMAP_RADIUS) continue;
          scores[ny][nx] += weight * (1 - dist / (HEATMAP_RADIUS + 1));
        }
      }
    }
  }

  // Normalize to [0, 1]
  let maxScore = 0;
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (scores[y][x] > maxScore) maxScore = scores[y][x];
    }
  }

  if (maxScore > 0) {
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        scores[y][x] /= maxScore;
      }
    }
  }

  return scores;
}

/**
 * Map a normalized threat score [0,1] to a CSS color string.
 * Gradient: transparent → blue → cyan → green → orange → red
 */
export function heatmapScoreToColor(score: number): string {
  if (score < 0.01) return 'transparent';

  type Stop = [number, [number, number, number], number];
  const stops: Stop[] = [
    [0.0,  [0,   0,   255], 0.0],
    [0.25, [0,   200, 255], 0.25],
    [0.5,  [0,   220, 80],  0.40],
    [0.75, [255, 165, 0],   0.55],
    [1.0,  [255, 0,   0],   0.70],
  ];

  let lower: Stop = stops[0];
  let upper: Stop = stops[stops.length - 1];
  for (let i = 0; i < stops.length - 1; i++) {
    if (score >= stops[i][0] && score <= stops[i + 1][0]) {
      lower = stops[i];
      upper = stops[i + 1];
      break;
    }
  }

  const range = upper[0] - lower[0];
  const t = range === 0 ? 0 : (score - lower[0]) / range;
  const r = Math.round(lower[1][0] + t * (upper[1][0] - lower[1][0]));
  const g = Math.round(lower[1][1] + t * (upper[1][1] - lower[1][1]));
  const b = Math.round(lower[1][2] + t * (upper[1][2] - lower[1][2]));
  const a = (lower[2] + t * (upper[2] - lower[2])).toFixed(2);

  return `rgba(${r},${g},${b},${a})`;
}
