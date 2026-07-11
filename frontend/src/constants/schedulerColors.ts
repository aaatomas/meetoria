import type { SchedulerEventColor } from '@mui/x-scheduler/models';

export const SERVICE_EVENT_COLORS: SchedulerEventColor[] = [
  'red',
  'pink',
  'purple',
  'indigo',
  'blue',
  'teal',
  'green',
  'lime',
  'amber',
  'orange',
  'grey',
];

export const DEFAULT_SERVICE_EVENT_COLOR: SchedulerEventColor = 'teal';

export function isServiceEventColor(value: string | undefined): value is SchedulerEventColor {
  return SERVICE_EVENT_COLORS.includes(value as SchedulerEventColor);
}

export function normalizeServiceEventColor(value: string | undefined): SchedulerEventColor {
  return isServiceEventColor(value) ? value : DEFAULT_SERVICE_EVENT_COLOR;
}
