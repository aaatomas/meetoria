import { createTheme } from '@mui/material/styles';
import '@mui/x-scheduler/theme-augmentation';

const solidCalendarEventStyles = {
  backgroundColor: 'var(--event-surface-selected) !important',
  color: 'var(--event-on-surface-selected) !important',
  '&:hover': {
    backgroundColor: 'var(--event-surface-selected-hover) !important',
  },
};

const solidCalendarEventTextStyles = {
  color: 'var(--event-on-surface-selected) !important',
};

const calendarEventRailStyles = {
  '&::before': {
    content: '""',
    position: 'absolute',
    top: 0,
    bottom: 0,
    left: 0,
    width: '4px !important',
    borderRadius: '4px 0 0 4px !important',
    background: 'color-mix(in srgb, var(--event-on-surface-selected) 35%, transparent) !important',
    pointerEvents: 'none',
  },
};

export const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#6C3AED',
      light: '#8B5CF6',
      dark: '#5B21B6',
    },
    secondary: {
      main: '#EC4899',
    },
    background: {
      default: '#F8FAFC',
      paper: '#FFFFFF',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h4: { fontWeight: 700 },
    h5: { fontWeight: 600 },
    h6: { fontWeight: 600 },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        html: { height: '100%' },
        body: { height: '100%', margin: 0 },
        '#root': { height: '100%' },
        '.MuiEventDialog-root, .MuiEventCalendar-eventDialog': {
          display: 'none !important',
          pointerEvents: 'none',
        },
        '.MuiEventCalendar-timeGridEvent:not([data-editing])': {
          ...solidCalendarEventStyles,
          ...calendarEventRailStyles,
        },
        '.MuiEventCalendar-timeGridEvent:not([data-editing]) .MuiEventCalendar-timeGridEventTitle': solidCalendarEventTextStyles,
        '.MuiEventCalendar-timeGridEvent:not([data-editing]) .MuiEventCalendar-timeGridEventTime': solidCalendarEventTextStyles,
        '.MuiEventCalendar-dayGridEvent[data-variant="filled"]:not([data-editing])': {
          ...solidCalendarEventStyles,
          position: 'relative',
          paddingLeft: '12px !important',
          ...calendarEventRailStyles,
        },
        '.MuiEventCalendar-dayGridEvent[data-variant="filled"]:not([data-editing]) .MuiEventCalendar-dayGridEventTitle': solidCalendarEventTextStyles,
        '.MuiEventCalendar-dayGridEvent[data-variant="filled"]:not([data-editing]) .MuiEventCalendar-dayGridEventTime': solidCalendarEventTextStyles,
        '.MuiEventCalendar-eventItemCard[data-variant="filled"]:not([data-editing])': {
          ...solidCalendarEventStyles,
          position: 'relative',
          paddingLeft: '12px !important',
          ...calendarEventRailStyles,
        },
        '.MuiEventCalendar-eventItemCard[data-variant="filled"]:not([data-editing]) .MuiEventCalendar-eventItemTitle': solidCalendarEventTextStyles,
        '.MuiEventCalendar-eventItemCard[data-variant="filled"]:not([data-editing]) .MuiEventCalendar-eventItemTime': solidCalendarEventTextStyles,
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        },
      },
    },
    MuiEventCalendar: {
      styleOverrides: {
        root: ({ theme }) => ({
          backgroundColor: theme.palette.background.default,
        }),
        sidePanel: ({ theme }) => ({
          backgroundColor: theme.palette.background.default,
        }),
        resourcesTreeLabel: {
          fontWeight: 700,
          fontSize: '0.8125rem',
          letterSpacing: '0.02em',
          px: 1.5,
          pt: 1,
        },
        resourcesTreeItem: ({ theme }) => ({
          borderRadius: 8,
          mx: 0.75,
          '&:hover': {
            backgroundColor: theme.palette.action.hover,
          },
        }),
        resourcesTreeItemCheckbox: {
          p: 0.5,
        },
        dayTimeGridAllDayEventsCell: {
          display: 'none',
        },
        dayTimeGridAllDayEventsGrid: {
          display: 'none',
        },
      },
    },
  },
});
