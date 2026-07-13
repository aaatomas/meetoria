import { useEffect, useState, type ReactNode } from 'react';
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  InputLabel,
  Link,
  MenuItem,
  Select,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material';
import { TimePicker } from '@mui/x-date-pickers/TimePicker';
import { useForm, Controller, useWatch } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  api,
  getApiErrorMessage,
  parseOrganizationSettings,
  type BookingSettings,
  type DaySchedule,
  type Organization,
  dayName,
  defaultWeekSchedule,
  getWorkingHours,
  saveWorkingHours,
} from '../../api/client';
import { parseTimeInputValue, formatTimeForStorage, timePickerProps, type TimeFormat } from '../../utils/dateTimeFieldUtils';
import { findTimezoneOption, getTimezoneOptions } from '../../utils/timezones';

const timezoneOptions = getTimezoneOptions();

const generalSchema = z.object({
  currency: z.string().length(3),
  timezone: z.string().min(1),
  time_format: z.enum(['24h', '12h']),
  email: z.string().email().optional().or(z.literal('')),
  slug: z.string().min(2).regex(/^[a-z0-9-]+$/),
});

type GeneralForm = z.infer<typeof generalSchema>;

const bookingSchema = z.object({
  enabled: z.boolean(),
  booking_window_days: z.coerce.number().min(1).max(365),
  min_notice_minutes: z.coerce.number().min(0),
  max_notice_minutes: z.coerce.number().min(0).optional().or(z.literal('')),
  email_required: z.boolean(),
  auto_confirm: z.boolean(),
  manual_approval: z.boolean(),
  cancellation_policy: z.string().optional(),
  rescheduling_policy: z.string().optional(),
});

type BookingForm = z.infer<typeof bookingSchema>;

function TabPanel({ children, value, index }: { children: ReactNode; value: number; index: number }) {
  if (value !== index) return null;
  return <Box pt={2}>{children}</Box>;
}

function WorkingHoursEditor({
  schedule,
  onChange,
  timeFormat,
}: {
  schedule: DaySchedule[];
  onChange: (schedule: DaySchedule[]) => void;
  timeFormat: TimeFormat;
}) {
  const pickerProps = timePickerProps(timeFormat);

  const updateDay = (dayOfWeek: number, patch: Partial<DaySchedule>) => {
    onChange(schedule.map((day) => (day.day_of_week === dayOfWeek ? { ...day, ...patch } : day)));
  };

  return (
    <Stack spacing={1.5}>
      {schedule.map((day) => {
        const open = day.slots.length > 0;
        const slot = day.slots[0] ?? { start_time: '09:00', end_time: '17:00' };
        return (
          <Box
            key={day.day_of_week}
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: '1fr',
                sm: '168px minmax(0, 1fr) minmax(0, 1fr)',
              },
              gap: 1.5,
              alignItems: 'center',
            }}
          >
            <FormControlLabel
              sx={{ m: 0 }}
              control={
                <Switch
                  checked={open}
                  onChange={(e) =>
                    updateDay(day.day_of_week, {
                      slots: e.target.checked ? [{ start_time: '09:00', end_time: '17:00' }] : [],
                    })
                  }
                />
              }
              label={dayName(day.day_of_week)}
            />
            {open ? (
              <>
                <TimePicker
                  label="From"
                  {...pickerProps}
                  views={['hours', 'minutes']}
                  value={parseTimeInputValue(slot.start_time)}
                  onChange={(next) => {
                    if (!next) return;
                    updateDay(day.day_of_week, {
                      slots: [{ start_time: formatTimeForStorage(next), end_time: slot.end_time }],
                    });
                  }}
                  slotProps={{ textField: { size: 'small', fullWidth: true } }}
                />
                <TimePicker
                  label="To"
                  {...pickerProps}
                  views={['hours', 'minutes']}
                  value={parseTimeInputValue(slot.end_time)}
                  onChange={(next) => {
                    if (!next) return;
                    updateDay(day.day_of_week, {
                      slots: [{ start_time: slot.start_time, end_time: formatTimeForStorage(next) }],
                    });
                  }}
                  slotProps={{ textField: { size: 'small', fullWidth: true } }}
                />
              </>
            ) : (
              <>
                <Box sx={{ display: { xs: 'none', sm: 'block' } }} />
                <Box sx={{ display: { xs: 'none', sm: 'block' } }} />
              </>
            )}
          </Box>
        );
      })}
    </Stack>
  );
}

interface OrganizationSettingsDialogProps {
  org: Organization | null;
  open: boolean;
  onClose: () => void;
}

export function OrganizationSettingsDialog({ org, open, onClose }: OrganizationSettingsDialogProps) {
  const queryClient = useQueryClient();
  const [schedule, setSchedule] = useState<DaySchedule[]>(defaultWeekSchedule());
  const [tab, setTab] = useState(0);

  useEffect(() => {
    if (open) setTab(0);
  }, [open, org?.id]);

  const orgQuery = useQuery({
    queryKey: ['organization', org?.id],
    queryFn: async () => {
      const { data } = await api.get<Organization>(`/organizations/${org!.id}`);
      return data;
    },
    enabled: open && !!org?.id,
  });

  useQuery({
    queryKey: ['working-hours', org?.id],
    queryFn: async () => {
      const data = await getWorkingHours(org!.id);
      setSchedule(data);
      return data;
    },
    enabled: open && !!org?.id,
  });

  const { control: generalControl, handleSubmit: handleGeneralSubmit, reset: resetGeneral } = useForm<GeneralForm>({
    resolver: zodResolver(generalSchema),
    defaultValues: { currency: 'EUR', timezone: 'UTC', time_format: '24h', email: '', slug: '' },
  });

  const timeFormat = useWatch({ control: generalControl, name: 'time_format' }) ?? '24h';

  const { control, handleSubmit, reset } = useForm<BookingForm>({
    resolver: zodResolver(bookingSchema),
    defaultValues: {
      enabled: true,
      booking_window_days: 30,
      min_notice_minutes: 60,
      max_notice_minutes: '',
      email_required: false,
      auto_confirm: true,
      manual_approval: false,
      cancellation_policy: '',
      rescheduling_policy: '',
    },
  });

  useEffect(() => {
    if (orgQuery.data) {
      const settings = parseOrganizationSettings(orgQuery.data.settings);
      resetGeneral({
        currency: orgQuery.data.currency || 'EUR',
        timezone: orgQuery.data.timezone || 'UTC',
        time_format: settings.time_format,
        email: orgQuery.data.email ?? '',
        slug: orgQuery.data.slug,
      });
      reset({
        enabled: settings.booking.enabled,
        booking_window_days: settings.booking.booking_window_days,
        min_notice_minutes: settings.booking.min_notice_minutes,
        max_notice_minutes: settings.booking.max_notice_minutes ?? '',
        email_required: settings.booking.email_required,
        auto_confirm: settings.booking.auto_confirm,
        manual_approval: settings.booking.manual_approval,
        cancellation_policy: settings.booking.cancellation_policy ?? '',
        rescheduling_policy: settings.booking.rescheduling_policy ?? '',
      });
    }
  }, [orgQuery.data, reset, resetGeneral]);

  const updateGeneral = useMutation({
    mutationFn: (data: GeneralForm) =>
      api.put(`/organizations/${org!.id}`, {
        currency: data.currency.toUpperCase(),
        timezone: data.timezone,
        time_format: data.time_format,
        email: data.email || '',
        slug: data.slug,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization', org?.id] });
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
      queryClient.invalidateQueries({ queryKey: ['services'] });
    },
  });

  const updateBooking = useMutation({
    mutationFn: (data: BookingForm) => {
      const booking: BookingSettings = {
        enabled: data.enabled,
        booking_window_days: data.booking_window_days,
        min_notice_minutes: data.min_notice_minutes,
        email_required: data.email_required,
        auto_confirm: data.auto_confirm,
        manual_approval: data.manual_approval,
        cancellation_policy: data.cancellation_policy,
        rescheduling_policy: data.rescheduling_policy,
      };
      if (data.max_notice_minutes !== '' && data.max_notice_minutes != null && Number(data.max_notice_minutes) > 0) {
        booking.max_notice_minutes = Number(data.max_notice_minutes);
      }
      return api.put(`/organizations/${org!.id}`, { booking });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization', org?.id] });
    },
  });

  const saveWorkingHoursMutation = useMutation({
    mutationFn: () => saveWorkingHours(org!.id, schedule),
    onSuccess: (data) => {
      setSchedule(data);
      queryClient.invalidateQueries({ queryKey: ['working-hours', org?.id] });
    },
  });

  const isSaving = updateGeneral.isPending || updateBooking.isPending || saveWorkingHoursMutation.isPending;
  const saveSuccess = updateGeneral.isSuccess || updateBooking.isSuccess || saveWorkingHoursMutation.isSuccess;

  const onSave = () => {
    if (tab === 0) {
      handleGeneralSubmit((d) => updateGeneral.mutate(d))();
    } else if (tab === 1) {
      saveWorkingHoursMutation.mutate();
    } else {
      handleSubmit((d) => updateBooking.mutate(d))();
    }
  };

  const publicBookingUrl = orgQuery.data?.slug
    ? `${window.location.origin}/book/${orgQuery.data.slug}`
    : null;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>{org?.name ?? 'Organization'} settings</DialogTitle>
      <DialogContent dividers sx={{ minHeight: 360 }}>
        {orgQuery.isLoading && <Typography color="text.secondary">Loading…</Typography>}
        {orgQuery.data && (
          <>
            <Tabs value={tab} onChange={(_, value) => setTab(value)}>
              <Tab label="General" />
              <Tab label="Working Hours" />
              <Tab label="Public Booking" />
              <Tab label="Policies" />
            </Tabs>

            <TabPanel value={tab} index={0}>
              <Typography variant="body2" color="text.secondary" mb={2}>
                General business details used across the app and public booking page.
              </Typography>
              <Stack spacing={2}>
                <Controller
                  name="slug"
                  control={generalControl}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="URL Slug"
                      fullWidth
                      helperText={
                        publicBookingUrl
                          ? `Public booking: ${publicBookingUrl}`
                          : 'Used in your public booking link'
                      }
                      onChange={(e) => field.onChange(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                    />
                  )}
                />
                <Controller
                  name="timezone"
                  control={generalControl}
                  render={({ field }) => (
                    <Autocomplete
                      options={timezoneOptions}
                      value={field.value ? findTimezoneOption(field.value) : null}
                      onChange={(_, option) => field.onChange(option?.value ?? '')}
                      getOptionLabel={(option) => option.label}
                      isOptionEqualToValue={(option, value) => option.value === value.value}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Timezone"
                          fullWidth
                          helperText="Search by city or UTC offset"
                        />
                      )}
                    />
                  )}
                />
                <Controller
                  name="time_format"
                  control={generalControl}
                  render={({ field }) => (
                    <FormControl fullWidth>
                      <InputLabel>Time format</InputLabel>
                      <Select {...field} label="Time format">
                        <MenuItem value="24h">24-hour (13:30)</MenuItem>
                        <MenuItem value="12h">12-hour (1:30 PM)</MenuItem>
                      </Select>
                    </FormControl>
                  )}
                />
                <Controller
                  name="email"
                  control={generalControl}
                  render={({ field }) => (
                    <TextField {...field} label="Business Email" type="email" fullWidth />
                  )}
                />
                <Controller
                  name="currency"
                  control={generalControl}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Currency"
                      fullWidth
                      inputProps={{ maxLength: 3, style: { textTransform: 'uppercase' } }}
                      helperText="ISO 4217 code, e.g. EUR, USD, GBP"
                      onChange={(e) => field.onChange(e.target.value.toUpperCase())}
                    />
                  )}
                />
                {updateGeneral.isError && (
                  <Alert severity="error">{getApiErrorMessage(updateGeneral.error)}</Alert>
                )}
                {saveSuccess && tab === 0 && <Alert severity="success">Settings saved.</Alert>}
              </Stack>
            </TabPanel>

            <TabPanel value={tab} index={1}>
              <Typography variant="body2" color="text.secondary" mb={2}>
                When customers can book appointments. Default is Mon–Fri 9:00–17:00.
              </Typography>
              <WorkingHoursEditor schedule={schedule} onChange={setSchedule} timeFormat={timeFormat} />
              {saveSuccess && tab === 1 && <Alert severity="success" sx={{ mt: 2 }}>Settings saved.</Alert>}
            </TabPanel>

            <TabPanel value={tab} index={2}>
              {publicBookingUrl && (
                <Alert severity="info" sx={{ mb: 2 }}>
                  Public booking:{' '}
                  <Link href={publicBookingUrl} target="_blank" rel="noopener">
                    {publicBookingUrl}
                  </Link>
                </Alert>
              )}
              <form id="org-booking-settings" onSubmit={handleSubmit((d) => updateBooking.mutate(d))}>
                <Stack spacing={2}>
                  <Controller
                    name="enabled"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={<Switch checked={field.value} onChange={field.onChange} />}
                        label="Enable public booking"
                      />
                    )}
                  />
                  <Controller
                    name="booking_window_days"
                    control={control}
                    render={({ field }) => (
                      <TextField {...field} label="Booking window (days ahead)" type="number" fullWidth />
                    )}
                  />
                  <Controller
                    name="min_notice_minutes"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        label="Minimum notice (minutes)"
                        type="number"
                        fullWidth
                        helperText="How soon before an appointment customers can book"
                      />
                    )}
                  />
                  <Controller
                    name="max_notice_minutes"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        label="Maximum notice (minutes, optional)"
                        type="number"
                        fullWidth
                        helperText="Leave empty for no extra limit"
                      />
                    )}
                  />
                  <Controller
                    name="email_required"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={<Switch checked={field.value} onChange={field.onChange} />}
                        label="Require customer email"
                      />
                    )}
                  />
                  <Controller
                    name="auto_confirm"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={<Switch checked={field.value} onChange={field.onChange} />}
                        label="Auto-confirm bookings"
                      />
                    )}
                  />
                  <Controller
                    name="manual_approval"
                    control={control}
                    render={({ field }) => (
                      <FormControlLabel
                        control={<Switch checked={field.value} onChange={field.onChange} />}
                        label="Require manual approval (pending status)"
                      />
                    )}
                  />
                  {saveSuccess && tab === 2 && <Alert severity="success">Settings saved.</Alert>}
                </Stack>
              </form>
            </TabPanel>

            <TabPanel value={tab} index={3}>
              <Typography variant="body2" color="text.secondary" mb={2}>
                Shown to customers on the public booking confirmation step. These are informational only and not
                enforced automatically.
              </Typography>
              <Stack spacing={2} component="form" id="org-booking-policies">
                <Controller
                  name="cancellation_policy"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Cancellation policy"
                      multiline
                      rows={4}
                      fullWidth
                      placeholder="e.g. Free cancellation up to 24 hours before your appointment."
                    />
                  )}
                />
                <Controller
                  name="rescheduling_policy"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      label="Rescheduling policy"
                      multiline
                      rows={4}
                      fullWidth
                      placeholder="e.g. Contact us at least 12 hours in advance to reschedule."
                    />
                  )}
                />
                {saveSuccess && tab === 3 && <Alert severity="success">Settings saved.</Alert>}
              </Stack>
            </TabPanel>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
        {orgQuery.data && (
          <Button variant="contained" disabled={isSaving} onClick={onSave}>
            Save Settings
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
}
