import { alpha, useTheme } from '@mui/material/styles';
import { Box, Typography } from '@mui/material';
import { UnfoldMore } from '@mui/icons-material';
import type { DashboardScope } from '../../api/client';

interface DashboardScopeBadgeProps {
  scope: DashboardScope;
  label: string;
  disabled?: boolean;
  onClick: () => void;
}

export function DashboardScopeBadge({ scope, label, disabled, onClick }: DashboardScopeBadgeProps) {
  const theme = useTheme();
  const isBranch = scope === 'branch';

  return (
    <Box
      component="button"
      type="button"
      onClick={onClick}
      disabled={disabled}
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 0.25,
        height: 28,
        px: 1.25,
        border: '1px solid',
        borderColor: isBranch
          ? alpha(theme.palette.primary.main, 0.35)
          : theme.palette.divider,
        borderRadius: 1.5,
        bgcolor: isBranch
          ? alpha(theme.palette.primary.main, 0.08)
          : theme.palette.background.paper,
        color: isBranch ? 'primary.main' : 'text.secondary',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.55 : 1,
        font: 'inherit',
        transition: 'border-color 0.15s ease, background-color 0.15s ease, box-shadow 0.15s ease',
        boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
        '&:hover': disabled
          ? {}
          : {
              borderColor: isBranch ? 'primary.main' : alpha(theme.palette.text.primary, 0.2),
              bgcolor: isBranch
                ? alpha(theme.palette.primary.main, 0.12)
                : alpha(theme.palette.action.hover, 0.35),
              boxShadow: '0 1px 3px rgba(15, 23, 42, 0.08)',
            },
      }}
    >
      <Typography
        component="span"
        variant="body2"
        fontWeight={600}
        noWrap
        sx={{ fontSize: '0.8125rem', lineHeight: 1.2, maxWidth: 180 }}
      >
        {label}
      </Typography>
      <UnfoldMore sx={{ fontSize: 15, opacity: 0.55, ml: -0.25, flexShrink: 0 }} />
    </Box>
  );
}
