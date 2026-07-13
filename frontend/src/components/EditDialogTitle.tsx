import { DialogTitle, FormControlLabel, Switch } from '@mui/material';
import { Controller, type Control, type FieldPath, type FieldValues } from 'react-hook-form';

interface EditDialogTitleProps<T extends FieldValues> {
  title: string;
  showActive?: boolean;
  activeDisabled?: boolean;
  control: Control<T>;
}

export function EditDialogTitle<T extends FieldValues>({
  title,
  showActive = false,
  activeDisabled = false,
  control,
}: EditDialogTitleProps<T>) {
  return (
    <DialogTitle
      component="div"
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 2,
        pr: 2,
      }}
    >
      {title}
      {showActive && (
        <Controller
          name={'is_active' as FieldPath<T>}
          control={control}
          render={({ field }) => (
            <FormControlLabel
              control={
                <Switch
                  checked={Boolean(field.value)}
                  onChange={field.onChange}
                  disabled={activeDisabled}
                />
              }
              label="Active"
              labelPlacement="start"
              sx={{ mr: 0, flexShrink: 0 }}
            />
          )}
        />
      )}
    </DialogTitle>
  );
}
