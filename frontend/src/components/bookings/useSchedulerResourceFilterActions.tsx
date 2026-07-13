import { RefObject, useLayoutEffect, useRef } from 'react';
import { createRoot, Root } from 'react-dom/client';
import { alpha, useTheme } from '@mui/material/styles';
import { Box, Button, ButtonGroup, Chip, Stack, Typography } from '@mui/material';
import { DoneAll, RemoveDone } from '@mui/icons-material';

const RESOURCES_TREE_SELECTOR = '.MuiEventCalendar-resourcesTree';
const RESOURCE_ACTIONS_ATTR = 'data-meetoria-resource-actions';

interface ResourceFilterStats {
  selectedCount: number;
  totalCount: number;
}

interface ResourceFilterActionsProps {
  onSelectAll: () => void;
  onDeselectAll: () => void;
  stats: ResourceFilterStats;
}

function ResourceFilterActions({ onSelectAll, onDeselectAll, stats }: ResourceFilterActionsProps) {
  const theme = useTheme();
  const { selectedCount, totalCount } = stats;
  const allSelected = totalCount > 0 && selectedCount === totalCount;
  const noneSelected = selectedCount === 0;

  return (
    <Box
      sx={{
        mx: 1.25,
        mb: 1.25,
        p: 1,
        borderRadius: 2,
        border: `1px solid ${theme.palette.divider}`,
        bgcolor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.04),
      }}
    >
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={1}>
        <Typography variant="caption" color="text.secondary" fontWeight={600}>
          Visible on calendar
        </Typography>
        <Chip
          size="small"
          label={`${selectedCount}/${totalCount}`}
          color={allSelected ? 'primary' : 'default'}
          variant={allSelected ? 'filled' : 'outlined'}
          sx={{ height: 22, fontSize: '0.7rem', fontWeight: 700 }}
        />
      </Stack>

      <ButtonGroup
        fullWidth
        variant="outlined"
        size="small"
        sx={{
          '& .MuiButton-root': {
            textTransform: 'none',
            fontWeight: 600,
            fontSize: '0.8125rem',
            py: 0.625,
            borderColor: theme.palette.divider,
            bgcolor: theme.palette.background.paper,
            '&:hover': {
              bgcolor: alpha(theme.palette.primary.main, 0.08),
              borderColor: theme.palette.primary.main,
            },
          },
        }}
      >
        <Button
          onClick={onSelectAll}
          disabled={allSelected || totalCount === 0}
          startIcon={<DoneAll sx={{ fontSize: 16 }} />}
        >
          Select all
        </Button>
        <Button
          onClick={onDeselectAll}
          disabled={noneSelected || totalCount === 0}
          startIcon={<RemoveDone sx={{ fontSize: 16 }} />}
        >
          Clear all
        </Button>
      </ButtonGroup>
    </Box>
  );
}

export function useSchedulerResourceFilterActions(
  containerRef: RefObject<HTMLElement | null>,
  onSelectAll: () => void,
  onDeselectAll: () => void,
  stats: ResourceFilterStats,
) {
  const rootsRef = useRef<Map<Element, Root>>(new Map());
  const hostsRef = useRef<Set<Element>>(new Set());
  const onSelectAllRef = useRef(onSelectAll);
  const onDeselectAllRef = useRef(onDeselectAll);
  const statsRef = useRef(stats);

  onSelectAllRef.current = onSelectAll;
  onDeselectAllRef.current = onDeselectAll;
  statsRef.current = stats;

  useLayoutEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const cleanupRoot = (host: Element) => {
      const root = rootsRef.current.get(host);
      if (root) {
        root.unmount();
        rootsRef.current.delete(host);
      }
      hostsRef.current.delete(host);
      host.remove();
    };

    const renderHost = (host: HTMLElement) => {
      let root = rootsRef.current.get(host);
      if (!root) {
        root = createRoot(host);
        rootsRef.current.set(host, root);
        hostsRef.current.add(host);
      }

      root.render(
        <ResourceFilterActions
          onSelectAll={() => onSelectAllRef.current()}
          onDeselectAll={() => onDeselectAllRef.current()}
          stats={statsRef.current}
        />,
      );
    };

    const decorateFilter = () => {
      const treeRoot = container.querySelector(RESOURCES_TREE_SELECTOR);
      if (!treeRoot) {
        rootsRef.current.forEach((root) => root.unmount());
        rootsRef.current.clear();
        hostsRef.current.clear();
        container.querySelectorAll(`[${RESOURCE_ACTIONS_ATTR}]`).forEach(cleanupRoot);
        return;
      }

      const label = treeRoot.querySelector('.MuiEventCalendar-resourcesTreeLabel');
      if (!label) {
        return;
      }

      let host = treeRoot.querySelector(`[${RESOURCE_ACTIONS_ATTR}]`) as HTMLElement | null;
      if (!host) {
        host = document.createElement('div');
        host.setAttribute(RESOURCE_ACTIONS_ATTR, 'true');
        label.insertAdjacentElement('afterend', host);
      }

      renderHost(host);
    };

    decorateFilter();

    const observer = new MutationObserver(decorateFilter);
    observer.observe(container, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
      rootsRef.current.forEach((root) => root.unmount());
      rootsRef.current.clear();
      hostsRef.current.clear();
      container.querySelectorAll(`[${RESOURCE_ACTIONS_ATTR}]`).forEach(cleanupRoot);
    };
  }, [containerRef, stats.selectedCount, stats.totalCount]);

  useLayoutEffect(() => {
    hostsRef.current.forEach((host) => {
      const root = rootsRef.current.get(host);
      root?.render(
        <ResourceFilterActions
          onSelectAll={() => onSelectAllRef.current()}
          onDeselectAll={() => onDeselectAllRef.current()}
          stats={statsRef.current}
        />,
      );
    });
  }, [stats.selectedCount, stats.totalCount]);
}

function buildAllVisible(resourceIds: string[], visible: boolean): Record<string, boolean> {
  return Object.fromEntries(resourceIds.map((id) => [id, visible]));
}

export function buildVisibleResourcesState(
  resourceIds: string[],
  current: Record<string, boolean> | undefined,
  visible: boolean,
): Record<string, boolean> {
  const next = { ...current };

  resourceIds.forEach((id) => {
    next[id] = visible;
  });

  Object.keys(next).forEach((id) => {
    if (!resourceIds.includes(id)) {
      delete next[id];
    }
  });

  return next;
}

export function mergeVisibleResources(
  resourceIds: string[],
  current: Record<string, boolean> | undefined,
): Record<string, boolean> {
  if (!current || Object.keys(current).length === 0) {
    return buildAllVisible(resourceIds, true);
  }

  const next = { ...current };

  resourceIds.forEach((id) => {
    if (!(id in next)) {
      next[id] = true;
    }
  });

  Object.keys(next).forEach((id) => {
    if (!resourceIds.includes(id)) {
      delete next[id];
    }
  });

  return next;
}

export { buildAllVisible };
