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
  showBranch?: boolean;
  compact?: boolean;
}

export function PopularServicesList({
  services,
  currency = 'EUR',
  showBranch = false,
  compact = false,
}: PopularServicesListProps) {
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
    <Stack spacing={compact ? 1.25 : 2}>
      {services.map((service, index) => {
        const colorKey = SERVICE_EVENT_COLORS[index % SERVICE_EVENT_COLORS.length];
        const barColor = COLOR_SWATCHES[colorKey];
        const share = service.count / maxCount;

        return (
          <Box key={`${service.branch_id}:${service.service_id}`}>
            <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: compact ? 1 : 1.5, mb: compact ? 0.5 : 0.75 }}>
              <Box
                sx={{
                  width: compact ? 20 : 24,
                  height: compact ? 20 : 24,
                  borderRadius: '50%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: compact ? 10 : 12,
                  fontWeight: 700,
                  flexShrink: 0,
                  bgcolor: alpha(barColor, 0.15),
                  color: barColor,
                }}
              >
                {index + 1}
              </Box>
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography variant="body2" fontWeight={600} noWrap sx={{ fontSize: compact ? '0.8125rem' : undefined }}>
                  {service.service_name}
                </Typography>
                {showBranch && service.branch_name && (
                  <Typography variant="caption" color="text.secondary" noWrap display="block">
                    {service.branch_name}
                  </Typography>
                )}
              </Box>
              <Typography
                variant={compact ? 'caption' : 'body2'}
                color="text.secondary"
                sx={{ flexShrink: 0, textAlign: 'right', lineHeight: 1.3 }}
              >
                {compact ? (
                  <>
                    {service.count}
                    <br />
                    <Box component="span" fontWeight={600} color="text.primary">
                      {formatPrice(service.revenue, currency)}
                    </Box>
                  </>
                ) : (
                  service.count
                )}
              </Typography>
              {!compact && (
                <Typography
                  variant="body2"
                  fontWeight={600}
                  sx={{ flexShrink: 0, minWidth: 72, textAlign: 'right' }}
                >
                  {formatPrice(service.revenue, currency)}
                </Typography>
              )}
            </Box>
            <Box
              sx={{
                ml: compact ? 3.5 : 4.5,
                height: compact ? 6 : 8,
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
