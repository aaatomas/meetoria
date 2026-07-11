import dayjs, { Dayjs } from 'dayjs';

export function syncNativeInputValue(input: HTMLInputElement, value: string) {
  const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
    HTMLInputElement.prototype,
    'value',
  )?.set;

  nativeInputValueSetter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

export function parseDateInputValue(value: string): Dayjs | null {
  if (!value) {
    return null;
  }

  const parsed = dayjs(value, 'YYYY-MM-DD', true);
  return parsed.isValid() ? parsed : null;
}

export function parseTimeInputValue(value: string): Dayjs | null {
  if (!value) {
    return null;
  }

  const [hours, minutes] = value.split(':');
  const parsed = dayjs()
    .hour(Number(hours))
    .minute(Number(minutes))
    .second(0)
    .millisecond(0);

  return parsed.isValid() ? parsed : null;
}

export function combineDateAndTime(date: Dayjs | null, time: Dayjs | null): Dayjs | null {
  if (!date || !time) {
    return null;
  }

  return date.hour(time.hour()).minute(time.minute()).second(0).millisecond(0);
}

export function mergeDatePart(current: Dayjs | null, nextDate: Dayjs | null): Dayjs | null {
  if (!nextDate) {
    return null;
  }

  const base = current ?? nextDate;
  return base.year(nextDate.year()).month(nextDate.month()).date(nextDate.date());
}

export function mergeTimePart(current: Dayjs | null, nextTime: Dayjs | null): Dayjs | null {
  if (!nextTime) {
    return null;
  }

  const base = current ?? nextTime;
  return base.hour(nextTime.hour()).minute(nextTime.minute()).second(0).millisecond(0);
}
