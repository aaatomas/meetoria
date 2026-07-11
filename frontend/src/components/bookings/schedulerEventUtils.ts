export const BOOKING_CLASS_PREFIX = 'meetoria-booking-';

export function bookingClassName(bookingId: string): string {
  return `${BOOKING_CLASS_PREFIX}${bookingId}`;
}

export function parseBookingIdFromElement(element: Element): string | null {
  const className = element.className;
  if (typeof className !== 'string') {
    return null;
  }

  const match = className.match(/meetoria-booking-([0-9a-f-]{36})/i);
  return match?.[1] ?? null;
}
