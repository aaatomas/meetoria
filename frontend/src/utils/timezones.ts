export interface TimezoneOption {
  value: string;
  label: string;
  offsetMinutes: number;
}

const FALLBACK_TIMEZONES = [
  'UTC',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Vilnius',
  'Europe/Warsaw',
  'Europe/Helsinki',
  'Europe/Athens',
  'Europe/Istanbul',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Toronto',
  'America/Sao_Paulo',
  'Asia/Dubai',
  'Asia/Kolkata',
  'Asia/Singapore',
  'Asia/Tokyo',
  'Australia/Sydney',
  'Pacific/Auckland',
];

function getTimezoneNames(): string[] {
  if (typeof Intl !== 'undefined' && 'supportedValuesOf' in Intl) {
    return (Intl as typeof Intl & { supportedValuesOf: (key: string) => string[] }).supportedValuesOf(
      'timeZone',
    );
  }
  return FALLBACK_TIMEZONES;
}

function getOffsetMinutes(timeZone: string, date = new Date()): number {
  try {
    const formatter = new Intl.DateTimeFormat('en-US', {
      timeZone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    });
    const parts = Object.fromEntries(
      formatter
        .formatToParts(date)
        .filter((part) => part.type !== 'literal')
        .map((part) => [part.type, part.value]),
    );
    const asUTC = Date.UTC(
      Number(parts.year),
      Number(parts.month) - 1,
      Number(parts.day),
      Number(parts.hour),
      Number(parts.minute),
      Number(parts.second),
    );
    return Math.round((asUTC - date.getTime()) / 60000);
  } catch {
    return 0;
  }
}

function formatOffset(minutes: number): string {
  const sign = minutes >= 0 ? '+' : '-';
  const absolute = Math.abs(minutes);
  const hours = Math.floor(absolute / 60);
  const mins = absolute % 60;
  return `UTC${sign}${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`;
}

function toTimezoneOption(value: string): TimezoneOption {
  const offsetMinutes = getOffsetMinutes(value);
  return {
    value,
    label: `(${formatOffset(offsetMinutes)}) ${value.replace(/_/g, ' ')}`,
    offsetMinutes,
  };
}

let cachedOptions: TimezoneOption[] | null = null;

export function getTimezoneOptions(): TimezoneOption[] {
  if (cachedOptions) return cachedOptions;

  cachedOptions = getTimezoneNames()
    .map(toTimezoneOption)
    .sort((a, b) => a.offsetMinutes - b.offsetMinutes || a.value.localeCompare(b.value));

  return cachedOptions;
}

export function findTimezoneOption(value: string): TimezoneOption {
  const match = getTimezoneOptions().find((option) => option.value === value);
  if (match) return match;
  return toTimezoneOption(value);
}
