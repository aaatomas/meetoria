import { Stack, Typography } from '@mui/material';
import { TrendingDown, TrendingFlat, TrendingUp } from '@mui/icons-material';
import type { MetricTrend } from '../../api/client';

function formatTrendLabel(trend: MetricTrend): string {
  if (trend.change_pct != null) {
    const sign = trend.change_pct > 0 ? '+' : '';
    return `${sign}${trend.change_pct.toFixed(0)}%`;
  }
  if (trend.change > 0) {
    return 'New';
  }
  if (trend.change < 0) {
    return `${trend.change}`;
  }
  return '0%';
}

function trendColor(trend: MetricTrend): 'success.main' | 'error.main' | 'text.secondary' {
  if (trend.change > 0) {
    return 'success.main';
  }
  if (trend.change < 0) {
    return 'error.main';
  }
  return 'text.secondary';
}

function TrendIcon({ trend }: { trend: MetricTrend }) {
  if (trend.change > 0) {
    return <TrendingUp sx={{ fontSize: 16 }} />;
  }
  if (trend.change < 0) {
    return <TrendingDown sx={{ fontSize: 16 }} />;
  }
  return <TrendingFlat sx={{ fontSize: 16 }} />;
}

interface StatTrendProps {
  trend?: MetricTrend;
}

export function StatTrend({ trend }: StatTrendProps) {
  if (!trend) {
    return null;
  }

  return (
    <Stack
      direction="row"
      alignItems="center"
      spacing={0.25}
      sx={{ color: trendColor(trend), flexShrink: 0 }}
    >
      <TrendIcon trend={trend} />
      <Typography variant="caption" fontWeight={600} lineHeight={1.2}>
        {formatTrendLabel(trend)}
      </Typography>
    </Stack>
  );
}
