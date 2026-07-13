import { Stack } from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { TimePicker } from '@mui/x-date-pickers/TimePicker';
import type { Dayjs } from 'dayjs';
import { mergeDatePart, mergeTimePart } from '../../utils/dateTimeFieldUtils';

const datePickerSlotProps = {
  textField: {
    fullWidth: true,
    size: 'small' as const,
  },
};

interface BookingDateFieldProps {
  value: Dayjs | null;
  onChange: (value: Dayjs | null) => void;
  label?: string;
  error?: string;
  disabled?: boolean;
  minDate?: Dayjs;
}

export function BookingDateField({
  value,
  onChange,
  label = 'Date',
  error,
  disabled = false,
  minDate,
}: BookingDateFieldProps) {
  return (
    <DatePicker
      label={label}
      value={value}
      disabled={disabled}
      minDate={minDate}
      onChange={onChange}
      slotProps={{
        textField: {
          ...datePickerSlotProps.textField,
          error: !!error,
          helperText: error,
        },
      }}
    />
  );
}

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
            ...datePickerSlotProps.textField,
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
