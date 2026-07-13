import { Box, Tooltip, Typography } from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { Fragment, useMemo } from 'react';
import type { HeatmapCell } from '../../api/client';

const SLOTS_PER_DAY = 12;
const SLOT_TICKS = [0, 3, 6, 9];
const LABEL_COL_WIDTH = 36;

function weekdayLabels(locale: string): string[] {
  const formatter = new Intl.DateTimeFormat(locale, { weekday: 'short' });
  const monday = new Date(2024, 0, 1);
  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(monday);
    date.setDate(monday.getDate() + index);
    return formatter.format(date);
  });
}

function formatHeatmapSlotRange(slot: number): string {
  const startHour = slot * 2;
  const endHour = (startHour + 2) % 24;
  const start = String(startHour).padStart(2, '0');
  const end = String(endHour).padStart(2, '0');
  return `${start}:00 – ${end}:00`;
}

function cellIntensity(count: number, maxCount: number, isDark: boolean): number | null {
  if (count <= 0 || maxCount <= 0) {
    return null;
  }
  return (isDark ? 0.42 : 0.28) + (count / maxCount) * (isDark ? 0.58 : 0.62);
}

interface BusiestHoursHeatmapProps {
  heatmapData: HeatmapCell[][];
  compact?: boolean;
}

export function BusiestHoursHeatmap({ heatmapData, compact = false }: BusiestHoursHeatmapProps) {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  const cells = useMemo(() => heatmapData ?? [], [heatmapData]);
  const maxCount = useMemo(
    () => Math.max(0, ...cells.flat().map((cell) => cell.count)),
    [cells],
  );
  const weekdays = useMemo(
    () => weekdayLabels(typeof navigator !== 'undefined' ? navigator.language : 'en'),
    [],
  );

  if (!cells.length || maxCount === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        No data yet
      </Typography>
    );
  }

  const labelColWidth = compact ? 28 : LABEL_COL_WIDTH;
  const slotTicks = compact ? [0, 6] : SLOT_TICKS;
  const emptyBackground = alpha(theme.palette.text.primary, isDark ? 0.06 : 0.04);
  const cellSx = {
    aspectRatio: '1',
    minHeight: compact ? 10 : 14,
    borderRadius: 0.5,
    border: `1px solid ${alpha(theme.palette.divider, 0.6)}`,
  };

  return (
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: `${labelColWidth}px repeat(${SLOTS_PER_DAY}, minmax(0, 1fr))`,
        gap: compact ? 0.25 : 0.5,
        alignItems: 'center',
        height: '100%',
      }}
    >
      <Box />
      {Array.from({ length: SLOTS_PER_DAY }, (_, slot) => (
        <Typography
          key={`tick-${slot}`}
          variant="caption"
          color="text.secondary"
          sx={{
            fontSize: compact ? 9 : 10,
            textAlign: 'center',
            visibility: slotTicks.includes(slot) ? 'visible' : 'hidden',
          }}
        >
          {String(slot * 2).padStart(2, '0')}
        </Typography>
      ))}

      {cells.map((row, weekday) => (
        <Fragment key={weekdays[weekday]}>
          <Typography
            variant="caption"
            color="text.secondary"
            sx={{ fontSize: compact ? 10 : 11, pr: 0.5, textAlign: 'right' }}
          >
            {weekdays[weekday]}
          </Typography>
          {row.map((cell, slot) => {
            const intensity = cellIntensity(cell.count, maxCount, isDark);
            const tooltip = `${weekdays[weekday]}, ${formatHeatmapSlotRange(slot)}: ${cell.count} booking${cell.count === 1 ? '' : 's'}`;

            return (
              <Tooltip key={`${weekday}-${slot}`} title={tooltip} arrow>
                <Box
                  sx={{
                    ...cellSx,
                    backgroundColor: intensity == null
                      ? emptyBackground
                      : alpha(theme.palette.primary.main, intensity),
                    cursor: cell.count > 0 ? 'default' : undefined,
                  }}
                />
              </Tooltip>
            );
          })}
        </Fragment>
      ))}
    </Box>
  );
}
