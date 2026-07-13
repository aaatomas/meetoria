import { useMemo, useState, type MouseEvent } from 'react';
import { alpha, useTheme } from '@mui/material/styles';
import {
  Box,
  Chip,
  ListItemIcon,
  ListItemText,
  ListSubheader,
  Menu,
  MenuItem,
  Stack,
  Typography,
} from '@mui/material';
import { Check, Place, UnfoldMore } from '@mui/icons-material';
import type { Branch, Organization } from '../../api/client';

export type LocationOption = {
  org: Organization;
  branch: Branch;
  key: string;
};

type LocationSelectProps = {
  options: LocationOption[];
  value: string;
  selected?: LocationOption;
  onChange: (value: string) => void;
};

function branchSubtitle(branch: Branch, orgName: string, multipleOrgs: boolean) {
  const location = [branch.city, branch.country].filter(Boolean).join(', ');
  if (multipleOrgs) {
    return location ? `${orgName} · ${location}` : orgName;
  }
  return location || undefined;
}

export function LocationSelect({ options, value, selected, onChange }: LocationSelectProps) {
  const theme = useTheme();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const multipleOrgs = new Set(options.map((option) => option.org.id)).size > 1;

  const groupedOptions = useMemo(() => {
    const groups: Array<{ org: Organization; items: LocationOption[] }> = [];

    options.forEach((option) => {
      const existing = groups.find((group) => group.org.id === option.org.id);
      if (existing) {
        existing.items.push(option);
      } else {
        groups.push({ org: option.org, items: [option] });
      }
    });

    return groups;
  }, [options]);

  const renderTriggerLabel = (branchName: string, orgName: string) =>
    multipleOrgs ? `${branchName} · ${orgName}` : branchName;

  const handleOpen = (event: MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleSelect = (key: string) => {
    handleClose();
    if (key !== value) {
      onChange(key);
    }
  };

  return (
    <>
      <Box
        component="button"
        type="button"
        onClick={handleOpen}
        aria-haspopup="listbox"
        aria-expanded={open}
        sx={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 0.75,
          height: 34,
          minWidth: 140,
          maxWidth: { xs: 180, sm: 240 },
          mr: 1,
          px: 1,
          pr: 0.75,
          border: '1px solid',
          borderColor: open ? 'primary.main' : 'divider',
          borderRadius: 1.5,
          bgcolor: open ? alpha(theme.palette.primary.main, 0.04) : 'background.paper',
          boxShadow: open
            ? `0 0 0 3px ${alpha(theme.palette.primary.main, 0.12)}`
            : '0 1px 2px rgba(15, 23, 42, 0.04)',
          color: 'text.primary',
          cursor: 'pointer',
          font: 'inherit',
          textAlign: 'left',
          transition: 'border-color 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease',
          '&:hover': {
            bgcolor: alpha(theme.palette.action.hover, 0.35),
            borderColor: alpha(theme.palette.text.primary, 0.2),
          },
        }}
      >
        {selected ? (
          <>
            <Box
              sx={{
                width: 22,
                height: 22,
                borderRadius: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                bgcolor: alpha(theme.palette.primary.main, 0.1),
                color: 'primary.main',
              }}
            >
              <Place sx={{ fontSize: 14 }} />
            </Box>
            <Typography
              variant="body2"
              fontWeight={600}
              noWrap
              sx={{ fontSize: '0.8125rem', lineHeight: 1.2, flex: 1, minWidth: 0 }}
            >
              {renderTriggerLabel(selected.branch.name, selected.org.name)}
            </Typography>
          </>
        ) : (
          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ fontSize: '0.8125rem', flex: 1, textAlign: 'left' }}
          >
            Location
          </Typography>
        )}
        <UnfoldMore
          sx={{
            fontSize: 18,
            color: 'text.secondary',
            opacity: 0.7,
            flexShrink: 0,
            transform: open ? 'rotate(180deg)' : 'none',
            transition: 'transform 0.15s ease',
          }}
        />
      </Box>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        slotProps={{
          paper: {
            sx: {
              mt: 1,
              borderRadius: 2,
              border: `1px solid ${theme.palette.divider}`,
              boxShadow: '0 10px 28px rgba(15, 23, 42, 0.12)',
              maxHeight: 360,
              minWidth: 280,
              py: 0.5,
            },
          },
          list: {
            dense: true,
            sx: { py: 0.5 },
          },
        }}
      >
        {groupedOptions.flatMap((group) => [
          ...(multipleOrgs
            ? [
                <ListSubheader
                  key={`header-${group.org.id}`}
                  disableSticky
                  sx={{
                    bgcolor: 'transparent',
                    lineHeight: 1.2,
                    py: 1,
                    px: 1.5,
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    letterSpacing: '0.04em',
                    textTransform: 'uppercase',
                    color: 'text.secondary',
                  }}
                >
                  {group.org.name}
                </ListSubheader>,
              ]
            : []),
          ...group.items.map(({ branch, key }) => {
            const isSelected = key === value;
            const subtitle = branchSubtitle(branch, group.org.name, multipleOrgs);

            return (
              <MenuItem
                key={key}
                selected={isSelected}
                onClick={() => handleSelect(key)}
                sx={{
                  mx: 0.75,
                  my: 0.25,
                  py: 0.75,
                  px: 1,
                  minHeight: 44,
                  borderRadius: 1.5,
                  gap: 1,
                  '&.Mui-selected': {
                    bgcolor: alpha(theme.palette.primary.main, 0.08),
                    '&:hover': {
                      bgcolor: alpha(theme.palette.primary.main, 0.12),
                    },
                  },
                }}
              >
                <ListItemIcon sx={{ minWidth: 32 }}>
                  <Box
                    sx={{
                      width: 28,
                      height: 28,
                      borderRadius: 1.25,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      bgcolor: isSelected
                        ? alpha(theme.palette.primary.main, 0.14)
                        : alpha(theme.palette.text.primary, 0.06),
                      color: isSelected ? 'primary.main' : 'text.secondary',
                    }}
                  >
                    <Place sx={{ fontSize: 16 }} />
                  </Box>
                </ListItemIcon>
                <ListItemText
                  primary={(
                    <Stack direction="row" alignItems="center" spacing={0.75} useFlexGap>
                      <Typography
                        component="span"
                        variant="body2"
                        fontWeight={isSelected ? 700 : 600}
                        noWrap
                        sx={{ fontSize: '0.8125rem' }}
                      >
                        {branch.name}
                      </Typography>
                      {branch.is_default && (
                        <Chip
                          label="Default"
                          size="small"
                          sx={{
                            height: 18,
                            fontSize: '0.625rem',
                            fontWeight: 700,
                            '& .MuiChip-label': { px: 0.75 },
                          }}
                        />
                      )}
                    </Stack>
                  )}
                  secondary={subtitle}
                  secondaryTypographyProps={{
                    fontSize: '0.7rem',
                    lineHeight: 1.2,
                    noWrap: true,
                  }}
                  sx={{ my: 0 }}
                />
                {isSelected && (
                  <Check sx={{ fontSize: 18, color: 'primary.main', flexShrink: 0 }} />
                )}
              </MenuItem>
            );
          }),
        ])}
      </Menu>
    </>
  );
}
