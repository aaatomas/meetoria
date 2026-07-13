import { RefObject, useLayoutEffect, useRef } from 'react';
import { createRoot, Root } from 'react-dom/client';
import { alpha, useTheme } from '@mui/material/styles';
import { Chip, IconButton, Stack, Tooltip } from '@mui/material';
import { pink } from '@mui/material/colors';
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
    <Stack
      direction="row"
      alignItems="center"
      spacing={0.25}
      component="span"
      sx={{ ml: 0.75, verticalAlign: 'middle' }}
    >
      <Chip
        size="small"
        label={`${selectedCount}/${totalCount}`}
        sx={{
          height: 20,
          fontSize: '0.65rem',
          fontWeight: 700,
          bgcolor: alpha(pink[500], theme.palette.mode === 'dark' ? 0.24 : 0.12),
          color: theme.palette.mode === 'dark' ? pink[200] : pink[700],
          '& .MuiChip-label': { px: 0.75 },
        }}
      />
      <Tooltip title="Select all">
        <span>
          <IconButton
            size="small"
            onClick={onSelectAll}
            disabled={allSelected || totalCount === 0}
            aria-label="Select all employees"
            sx={{ p: 0.375, color: 'text.secondary' }}
          >
            <DoneAll sx={{ fontSize: 16 }} />
          </IconButton>
        </span>
      </Tooltip>
      <Tooltip title="Clear all">
        <span>
          <IconButton
            size="small"
            onClick={onDeselectAll}
            disabled={noneSelected || totalCount === 0}
            aria-label="Clear all employees"
            sx={{ p: 0.375, color: 'text.secondary' }}
          >
            <RemoveDone sx={{ fontSize: 16 }} />
          </IconButton>
        </span>
      </Tooltip>
    </Stack>
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

      const label = treeRoot.querySelector('.MuiEventCalendar-resourcesTreeLabel') as HTMLElement | null;
      if (!label) {
        return;
      }

      label.style.display = 'inline-flex';
      label.style.alignItems = 'center';
      label.style.flexWrap = 'wrap';
      label.style.gap = '2px';

      let host = label.querySelector(`[${RESOURCE_ACTIONS_ATTR}]`) as HTMLElement | null;
      if (!host) {
        host = document.createElement('span');
        host.setAttribute(RESOURCE_ACTIONS_ATTR, 'true');
        label.appendChild(host);
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
