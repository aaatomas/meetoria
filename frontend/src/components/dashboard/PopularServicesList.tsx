import { Box, Stack, Typography } from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import {
  amber,
  blue,
  deepOrange,
  green,
  grey,
  indigo,
  lime,
  pink,
  purple,
  red,
  teal,
} from '@mui/material/colors';
import type { SchedulerEventColor } from '@mui/x-scheduler/models';
import type { PopularService } from '../../api/client';
import { formatPrice } from '../../utils/formatCurrency';
import { SERVICE_EVENT_COLORS } from '../../constants/schedulerColors';

const COLOR_SWATCHES: Record<SchedulerEventColor, string> = {
  red: red[500],
  pink: pink[400],
  purple: purple[500],
  indigo: indigo[500],
  blue: blue[500],
  teal: teal[500],
  green: green[600],
  lime: lime[700],
  amber: amber[700],
  orange: deepOrange[500],
  grey: grey[600],
};

interface PopularServicesListProps {
  services: PopularService[];
  currency?: string;
}

export function PopularServicesList({ services, currency = 'EUR' }: PopularServicesListProps) {
  const theme = useTheme();

  if (!services.length) {
    return (
      <Typography variant="body2" color="text.secondary">
        No data yet
      </Typography>
    );
  }

  const maxCount = Math.max(...services.map((service) => service.count), 1);

  return (
    <Stack spacing={2}>
      {services.map((service, index) => {
        const colorKey = SERVICE_EVENT_COLORS[index % SERVICE_EVENT_COLORS.length];
        const barColor = COLOR_SWATCHES[colorKey];
        const share = service.count / maxCount;

        return (
          <Box key={service.service_id}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 0.75 }}>
              <Box
                sx={{
                  width: 24,
                  height: 24,
                  borderRadius: '50%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 12,
                  fontWeight: 700,
                  flexShrink: 0,
                  bgcolor: alpha(barColor, 0.15),
                  color: barColor,
                }}
              >
                {index + 1}
              </Box>
              <Typography variant="body2" fontWeight={600} sx={{ flex: 1, minWidth: 0 }} noWrap>
                {service.service_name}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ flexShrink: 0 }}>
                {service.count}
              </Typography>
              <Typography
                variant="body2"
                fontWeight={600}
                sx={{ flexShrink: 0, minWidth: 72, textAlign: 'right' }}
              >
                {formatPrice(service.revenue, currency)}
              </Typography>
            </Box>
            <Box
              sx={{
                ml: 4.5,
                height: 8,
                borderRadius: 999,
                bgcolor: alpha(theme.palette.text.primary, theme.palette.mode === 'dark' ? 0.08 : 0.06),
                overflow: 'hidden',
              }}
            >
              <Box
                sx={{
                  width: `${Math.max(share * 100, 4)}%`,
                  height: '100%',
                  borderRadius: 999,
                  background: `linear-gradient(90deg, ${alpha(barColor, 0.85)}, ${barColor})`,
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
