import { Stack } from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { TimePicker } from '@mui/x-date-pickers/TimePicker';
import type { Dayjs } from 'dayjs';
import { mergeDatePart, mergeTimePart } from '../../utils/dateTimeFieldUtils';

interface BookingDateTimeFieldsProps {
  value: Dayjs | null;
  onChange: (value: Dayjs | null) => void;
  dateLabel?: string;
  timeLabel?: string;
  error?: string;
  disabled?: boolean;
}

export function BookingDateTimeFields({
  value,
  onChange,
  dateLabel = 'Date',
  timeLabel = 'Time',
  error,
  disabled = false,
}: BookingDateTimeFieldsProps) {
  return (
    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
      <DatePicker
        label={dateLabel}
        value={value}
        disabled={disabled}
        onChange={(nextDate) => onChange(mergeDatePart(value, nextDate))}
        slotProps={{
          textField: {
            fullWidth: true,
            size: 'small',
            error: !!error,
          },
        }}
      />
      <TimePicker
        label={timeLabel}
        value={value}
        disabled={disabled}
        ampm={false}
        format="HH:mm"
        views={['hours', 'minutes']}
        onChange={(nextTime) => onChange(mergeTimePart(value, nextTime))}
        slotProps={{
          textField: {
            fullWidth: true,
            size: 'small',
            error: !!error,
            helperText: error,
          },
        }}
      />
    </Stack>
  );
}
