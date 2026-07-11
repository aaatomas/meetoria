import { RefObject, useLayoutEffect, useRef } from 'react';
import { parseBookingIdFromElement } from './schedulerEventUtils';

const EVENT_SELECTOR = '.MuiEventCalendar-dayGridEvent, .MuiEventCalendar-timeGridEvent';

export function useSchedulerEventClick(
  containerRef: RefObject<HTMLElement | null>,
  onBookingClick: (bookingId: string) => void,
  enabled = true,
) {
  const onBookingClickRef = useRef(onBookingClick);
  onBookingClickRef.current = onBookingClick;

  useLayoutEffect(() => {
    if (!enabled) {
      return;
    }

    const container = containerRef.current;
    if (!container) {
      return;
    }

    const handleClick = (event: MouseEvent) => {
      const target = event.target as Element | null;
      const eventElement = target?.closest(EVENT_SELECTOR);
      if (!eventElement || eventElement.getAttribute('aria-hidden') === 'true') {
        return;
      }

      const bookingId = parseBookingIdFromElement(eventElement);
      if (!bookingId) {
        return;
      }

      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
      onBookingClickRef.current(bookingId);
    };

    container.addEventListener('click', handleClick, true);

    return () => {
      container.removeEventListener('click', handleClick, true);
    };
  }, [containerRef, enabled]);
}
