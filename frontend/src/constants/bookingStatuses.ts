import type { SchedulerEventColor } from '@mui/x-scheduler/models';
import { getSchedulerEventSurfaceColors } from './schedulerEventPalette';

export const BOOKING_STATUSES = [
  'pending',
  'confirmed',
  'in_progress',
  'completed',
  'cancelled',
  'no_show',
] as const;

export const BOOKING_STATUS_CHANGE_OPTIONS = [
  'completed',
  'cancelled',
] as const;

export type BookingStatus = (typeof BOOKING_STATUSES)[number];

export const BOOKING_STATUS_EVENT_COLORS: Record<BookingStatus, SchedulerEventColor> = {
  pending: 'amber',
  confirmed: 'blue',
  in_progress: 'indigo',
  completed: 'green',
  cancelled: 'grey',
  no_show: 'pink',
};

export function resolveBookingStatusEventColor(status: string): SchedulerEventColor {
  if (BOOKING_STATUSES.includes(status as BookingStatus)) {
    return BOOKING_STATUS_EVENT_COLORS[status as BookingStatus];
  }
  return 'teal';
}

export function formatBookingStatus(status: string): string {
  return status.replace(/_/g, ' ');
}

export function getBookingStatusBadgeColors(status: string) {
  return getSchedulerEventSurfaceColors(resolveBookingStatusEventColor(status));
}

export function canChangeBookingStatus(status: string): boolean {
  return status !== 'completed' && status !== 'cancelled';
}
