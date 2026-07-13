import { Box, Stack, Typography } from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { useMemo } from 'react';
import type { DayCount } from '../../api/client';

const WEEKDAY_KEYS = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'] as const;

function weekdayLabels(locale: string): string[] {
  const formatter = new Intl.DateTimeFormat(locale, { weekday: 'short' });
  const monday = new Date(2024, 0, 1);
  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(monday);
    date.setDate(monday.getDate() + index);
    return formatter.format(date);
  });
}

function normalizeDayCounts(days: DayCount[]) {
  const counts = new Map<string, number>();
  days.forEach(({ day, count }) => {
    counts.set(day.trim().toLowerCase(), count);
  });

  const labels = weekdayLabels(typeof navigator !== 'undefined' ? navigator.language : 'en');
  return WEEKDAY_KEYS.map((key, index) => ({
    day: labels[index],
    count: counts.get(key) ?? 0,
  }));
}

interface BusiestDaysChartProps {
  days: DayCount[];
  compact?: boolean;
}

export function BusiestDaysChart({ days, compact = false }: BusiestDaysChartProps) {
  const theme = useTheme();
  const chartDays = useMemo(() => normalizeDayCounts(days), [days]);
  const maxCount = Math.max(...chartDays.map((day) => day.count), 0);

  if (maxCount === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        No data yet
      </Typography>
    );
  }

  return (
    <Stack spacing={compact ? 0.875 : 1.5} sx={{ height: '100%', justifyContent: 'space-between' }}>
      {chartDays.map(({ day, count }) => {
        const share = count / maxCount;

        return (
          <Box key={day}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: compact ? 0.25 : 0.5 }}>
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{
                  width: compact ? 28 : 36,
                  flexShrink: 0,
                  fontSize: compact ? '0.75rem' : '0.8125rem',
                  fontWeight: 600,
                }}
              >
                {day}
              </Typography>
              <Typography
                variant="body2"
                fontWeight={600}
                sx={{ flex: 1, minWidth: 0, fontSize: compact ? '0.8125rem' : undefined }}
              >
                {count}
              </Typography>
            </Box>
            <Box
              sx={{
                ml: compact ? 4.5 : 5.5,
                height: compact ? 6 : 8,
                borderRadius: 999,
                bgcolor: alpha(theme.palette.text.primary, theme.palette.mode === 'dark' ? 0.08 : 0.06),
                overflow: 'hidden',
              }}
            >
              <Box
                sx={{
                  width: count > 0 ? `${Math.max(share * 100, 4)}%` : 0,
                  height: '100%',
                  borderRadius: 999,
                  background: `linear-gradient(90deg, ${alpha(theme.palette.secondary.main, 0.85)}, ${theme.palette.secondary.main})`,
                  transition: 'width 0.35s ease',
                }}
              />
            </Box>
          </Box>
        );
      })}
    </Stack>
  );
}
