import { useCallback, useEffect, type RefObject } from 'react';

const SCROLLABLE_SELECTORS = [
  '.MuiEventCalendar-dayTimeGridScrollableContent',
  '.MuiEventCalendar-dayTimeGrid',
] as const;
const NOW_INDICATOR_SELECTOR = '.MuiEventCalendar-dayTimeGridCurrentTimeIndicator';

function isScrollable(element: HTMLElement) {
  const { overflowY } = window.getComputedStyle(element);
  return (overflowY === 'auto' || overflowY === 'scroll')
    && element.scrollHeight > element.clientHeight + 1;
}

function findTimeGridScroller(root: HTMLElement, indicator?: HTMLElement | null) {
  if (indicator) {
    let parent = indicator.parentElement;
    while (parent && root.contains(parent)) {
      if (isScrollable(parent)) {
        return parent;
      }
      parent = parent.parentElement;
    }
  }

  for (const selector of SCROLLABLE_SELECTORS) {
    const candidate = root.querySelector(selector);
    if (candidate instanceof HTMLElement && isScrollable(candidate)) {
      return candidate;
    }
  }

  return null;
}

function scrollTimeGridToNow(root: HTMLElement) {
  const indicator = root.querySelector(NOW_INDICATOR_SELECTOR);
  if (!(indicator instanceof HTMLElement)) {
    return;
  }

  const scroller = findTimeGridScroller(root, indicator);
  if (!scroller) {
    return;
  }

  const scrollerRect = scroller.getBoundingClientRect();
  const indicatorRect = indicator.getBoundingClientRect();
  const delta = indicatorRect.top - scrollerRect.top - scroller.clientHeight / 2;
  scroller.scrollTop = Math.max(
    0,
    Math.min(scroller.scrollTop + delta, scroller.scrollHeight - scroller.clientHeight),
  );
}

function scheduleScrollAttempts(run: () => void) {
  run();
  window.requestAnimationFrame(() => {
    run();
    window.requestAnimationFrame(run);
  });
  window.setTimeout(run, 100);
  window.setTimeout(run, 300);
}

export function useSchedulerScrollToNow(
  schedulerRef: RefObject<HTMLDivElement | null>,
  enabled: boolean,
) {
  const scrollToNow = useCallback(() => {
    const root = schedulerRef.current;
    if (!root || !enabled) {
      return;
    }
    scrollTimeGridToNow(root);
  }, [enabled, schedulerRef]);

  useEffect(() => {
    if (!enabled) {
      return undefined;
    }

    let frame = 0;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    const scheduleScroll = () => {
      cancelAnimationFrame(frame);
      frame = window.requestAnimationFrame(() => {
        scheduleScrollAttempts(scrollToNow);
      });
    };

    const debouncedScroll = () => {
      clearTimeout(timeout);
      timeout = setTimeout(scheduleScroll, 50);
    };

    scheduleScroll();

    const root = schedulerRef.current;
    const observer = root
      ? new MutationObserver(debouncedScroll)
      : undefined;
    if (root && observer) {
      observer.observe(root, { childList: true, subtree: true });
    }

    const interval = window.setInterval(scheduleScroll, 60_000);

    return () => {
      cancelAnimationFrame(frame);
      clearTimeout(timeout);
      observer?.disconnect();
      clearInterval(interval);
    };
  }, [enabled, schedulerRef, scrollToNow]);

  return scrollToNow;
}
