import { TextField, type TextFieldProps } from '@mui/material';
import { formatPhoneDisplay, PHONE_FORMAT_MESSAGE, PHONE_PLACEHOLDER } from '../../utils/phoneUtils';

type PhoneFieldProps = Omit<TextFieldProps, 'type'> & {
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
};

export function PhoneField({
  value,
  onChange,
  onBlur,
  helperText,
  placeholder = PHONE_PLACEHOLDER,
  ...props
}: PhoneFieldProps) {
  return (
    <TextField
      {...props}
      type="tel"
      autoComplete="tel"
      value={value}
      placeholder={placeholder}
      helperText={helperText ?? PHONE_FORMAT_MESSAGE}
      onChange={(e) => onChange(e.target.value)}
      onBlur={() => {
        if (value.trim()) {
          onChange(formatPhoneDisplay(value));
        }
        onBlur?.();
      }}
    />
  );
}
