import {
  blue,
  deepOrange,
  grey,
  indigo,
  pink,
  purple,
  red,
  teal,
} from '@mui/material/colors';
import type { SchedulerEventColor } from '@mui/x-scheduler/models';

export interface SchedulerEventSurfaceColors {
  background: string;
  foreground: string;
  hoverBackground: string;
}

// Matches MUI X Scheduler light-mode event surface-selected tokens.
export const SCHEDULER_EVENT_SURFACE_COLORS: Record<SchedulerEventColor, SchedulerEventSurfaceColors> = {
  red: {
    background: red[600],
    foreground: '#FFFFFF',
    hoverBackground: red[700],
  },
  pink: {
    background: pink[400],
    foreground: '#FFFFFF',
    hoverBackground: '#e9437b',
  },
  purple: {
    background: purple[400],
    foreground: '#FFFFFF',
    hoverBackground: '#993ea8',
  },
  indigo: {
    background: indigo[500],
    foreground: '#FFFFFF',
    hoverBackground: indigo[600],
  },
  blue: {
    background: blue[700],
    foreground: '#FFFFFF',
    hoverBackground: blue[800],
  },
  teal: {
    background: teal[600],
    foreground: '#FFFFFF',
    hoverBackground: teal[700],
  },
  green: {
    background: '#2E7D32',
    foreground: '#FFFFFF',
    hoverBackground: '#316F35',
  },
  lime: {
    background: '#959B1F',
    foreground: '#FFFFFF',
    hoverBackground: '#898F1D',
  },
  amber: {
    background: '#e09312',
    foreground: '#FFFFFF',
    hoverBackground: '#E98300',
  },
  orange: {
    background: deepOrange[600],
    foreground: '#FFFFFF',
    hoverBackground: '#E85F14',
  },
  grey: {
    background: grey[600],
    foreground: '#FFFFFF',
    hoverBackground: grey[700],
  },
};

export function getSchedulerEventSurfaceColors(color: SchedulerEventColor): SchedulerEventSurfaceColors {
  return SCHEDULER_EVENT_SURFACE_COLORS[color];
}
