import { RefObject, useLayoutEffect, useRef } from 'react';
import { createRoot, Root } from 'react-dom/client';
import { Box, Button } from '@mui/material';

const RESOURCES_TREE_SELECTOR = '.MuiEventCalendar-resourcesTree';
const RESOURCE_ACTIONS_ATTR = 'data-meetoria-resource-actions';

interface ResourceFilterActionsProps {
  onSelectAll: () => void;
  onDeselectAll: () => void;
}

function ResourceFilterActions({ onSelectAll, onDeselectAll }: ResourceFilterActionsProps) {
  return (
    <Box
      sx={{
        display: 'flex',
        gap: 0.5,
        px: 1.5,
        pb: 1,
      }}
    >
      <Button size="small" variant="text" onClick={onSelectAll}>
        Select all
      </Button>
      <Button size="small" variant="text" onClick={onDeselectAll}>
        Clear all
      </Button>
    </Box>
  );
}

export function useSchedulerResourceFilterActions(
  containerRef: RefObject<HTMLElement | null>,
  onSelectAll: () => void,
  onDeselectAll: () => void,
) {
  const rootsRef = useRef<Map<Element, Root>>(new Map());
  const onSelectAllRef = useRef(onSelectAll);
  const onDeselectAllRef = useRef(onDeselectAll);

  onSelectAllRef.current = onSelectAll;
  onDeselectAllRef.current = onDeselectAll;

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
      host.remove();
    };

    const decorateFilter = () => {
      const treeRoot = container.querySelector(RESOURCES_TREE_SELECTOR);
      if (!treeRoot) {
        rootsRef.current.forEach((root) => root.unmount());
        rootsRef.current.clear();
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

      if (host.getAttribute('data-meetoria-decorated') === 'true' && rootsRef.current.has(host)) {
        return;
      }

      const existingRoot = rootsRef.current.get(host);
      if (existingRoot) {
        existingRoot.unmount();
        rootsRef.current.delete(host);
      }

      host.setAttribute('data-meetoria-decorated', 'true');

      const root = createRoot(host);
      rootsRef.current.set(host, root);
      root.render(
        <ResourceFilterActions
          onSelectAll={() => onSelectAllRef.current()}
          onDeselectAll={() => onDeselectAllRef.current()}
        />,
      );
    };

    decorateFilter();

    const observer = new MutationObserver(decorateFilter);
    observer.observe(container, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
      rootsRef.current.forEach((root) => root.unmount());
      rootsRef.current.clear();
      container.querySelectorAll(`[${RESOURCE_ACTIONS_ATTR}]`).forEach(cleanupRoot);
    };
  }, [containerRef]);
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
