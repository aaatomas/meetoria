import {
  Box,
  Typography,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  Stack,
  Alert,
} from '@mui/material';
import { Add } from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState, useMemo } from 'react';
import dayjs, { Dayjs } from 'dayjs';
import { api, Booking, Customer, PaginatedResponse, getActiveBranchId, getApiErrorMessage, listEmployees, listServices, parseOrganizationSettings, resolveActiveBranchId, type Organization } from '../api/client';
import { BookingScheduler } from '../components/bookings/BookingScheduler';
import { BookingDateTimeFields } from '../components/bookings/BookingDateTimeFields';
import { renderEmployeeLabel } from '../components/bookings/renderEmployeeLabel';

const bookingSchema = z.object({
  customer_id: z.string().min(1, 'Customer is required'),
  employee_id: z.string().min(1, 'Employee is required'),
  service_id: z.string().min(1, 'Service is required'),
  start_at: z.custom<Dayjs | null>(
    (value) => value !== null && dayjs.isDayjs(value) && value.isValid(),
    'Start date and time are required',
  ),
  notes: z.string().optional(),
});

type BookingForm = z.infer<typeof bookingSchema>;

const emptyOption = (label: string) => (
  <MenuItem value="" disabled>
    {label}
  </MenuItem>
);

export function BookingsPage() {
  const orgId = localStorage.getItem('organization_id')!;
  const branchId = getActiveBranchId();
  const [open, setOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: org } = useQuery({
    queryKey: ['organization', orgId],
    queryFn: async () => (await api.get<Organization>(`/organizations/${orgId}`)).data,
    enabled: !!orgId,
  });

  const currency = org?.currency?.trim() || 'EUR';
  const timeFormat = parseOrganizationSettings(org?.settings).time_format;

  const { data: bookings = [], isLoading: bookingsLoading } = useQuery({
    queryKey: ['bookings', orgId, branchId],
    queryFn: async () => {
      const { data } = await api.get<PaginatedResponse<Booking>>(`/organizations/${orgId}/bookings`, {
        params: { limit: 100, ...(branchId ? { branch_id: branchId } : {}) },
      });
      return data.data;
    },
    enabled: !!orgId && !!branchId,
  });

  const { data: customers = [] } = useQuery({
    queryKey: ['customers', orgId],
    queryFn: async () => (await api.get<PaginatedResponse<Customer>>(`/organizations/${orgId}/customers`)).data.data,
    enabled: !!orgId,
  });

  const { data: employees = [] } = useQuery({
    queryKey: ['employees', orgId, branchId],
    queryFn: () => listEmployees(orgId, branchId),
    enabled: !!orgId && !!branchId,
  });

  const { data: services = [] } = useQuery({
    queryKey: ['services', orgId, branchId],
    queryFn: () => listServices(orgId, branchId),
    enabled: !!orgId && !!branchId,
  });

  const activeEmployees = useMemo(
    () => employees.filter((employee) => employee.is_active),
    [employees],
  );

  const activeServices = useMemo(
    () => services.filter((service) => service.is_active),
    [services],
  );

  const { control, handleSubmit, reset, formState: { errors } } = useForm<BookingForm>({
    resolver: zodResolver(bookingSchema),
    defaultValues: {
      customer_id: '',
      employee_id: '',
      service_id: '',
      start_at: null,
      notes: '',
    },
  });

  const createMutation = useMutation({
    mutationFn: async (data: BookingForm) => {
      const activeBranchId = branchId ?? (await resolveActiveBranchId(orgId));
      if (!activeBranchId) {
        throw new Error('No location selected. Choose a location in the header.');
      }
      return api.post(`/organizations/${orgId}/bookings`, {
        customer_id: data.customer_id,
        employee_id: data.employee_id,
        service_id: data.service_id,
        branch_id: activeBranchId,
        start_time: data.start_at!.toISOString(),
        notes: data.notes ?? '',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
      setOpen(false);
      setSubmitError(null);
      reset();
    },
    onError: (error) => {
      setSubmitError(getApiErrorMessage(error));
    },
  });

  const canCreateBooking = customers.length > 0 && activeEmployees.length > 0 && activeServices.length > 0;

  const openDialog = () => {
    setSubmitError(null);
    reset();
    setOpen(true);
  };

  const onSubmit = handleSubmit(
    (data) => {
      setSubmitError(null);
      createMutation.mutate(data);
    },
    () => setSubmitError('Please fill in all required fields.'),
  );

  if (!orgId) {
    return (
      <Box>
        <Typography variant="h5" fontWeight={700} gutterBottom>Bookings</Typography>
        <Alert severity="info">Select or create an organization to manage bookings.</Alert>
      </Box>
    );
  }

  if (!branchId) {
    return (
      <Box>
        <Typography variant="h5" fontWeight={700} gutterBottom>Bookings</Typography>
        <Alert severity="info">Select a location in the header to view branch bookings.</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={1.5} flexShrink={0}>
        <Box>
          <Typography variant="h5" fontWeight={700}>Bookings</Typography>
          <Typography variant="body2" color="text.secondary">
            Drag appointments to reschedule. Click an event for details.
          </Typography>
        </Box>
        <Button variant="contained" startIcon={<Add />} onClick={openDialog}>
          New Booking
        </Button>
      </Box>

      {!canCreateBooking && (
        <Alert severity="info" sx={{ mb: 1.5, flexShrink: 0 }}>
          Add at least one customer, employee, and service before creating a booking.
        </Alert>
      )}

      <Box sx={{ flex: 1, minHeight: 0 }}>
        <BookingScheduler
          orgId={orgId}
          branchId={branchId}
          currency={currency}
          timeFormat={timeFormat}
          bookings={bookings}
          customers={customers}
          employees={employees}
          services={services}
          isLoading={bookingsLoading}
        />
      </Box>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>New Booking</DialogTitle>
        <Box component="form" onSubmit={onSubmit}>
          <DialogContent>
            <Stack spacing={2} mt={1}>
              {submitError && <Alert severity="error">{submitError}</Alert>}
              <Controller name="customer_id" control={control} render={({ field }) => (
                <TextField {...field} select label="Customer" fullWidth error={!!errors.customer_id} helperText={errors.customer_id?.message}>
                  {emptyOption('Select customer')}
                  {customers.map((c) => <MenuItem key={c.id} value={c.id}>{c.first_name} {c.last_name}</MenuItem>)}
                </TextField>
              )} />
              <Controller name="employee_id" control={control} render={({ field }) => (
                <TextField
                  {...field}
                  select
                  label="Employee"
                  fullWidth
                  error={!!errors.employee_id}
                  helperText={errors.employee_id?.message}
                  slotProps={{
                    select: {
                      renderValue: (value) => {
                        const employee = activeEmployees.find((item) => item.id === value);
                        return employee ? renderEmployeeLabel(employee, { avatarSize: 20, showInactive: false }) : '';
                      },
                    },
                  }}
                >
                  {emptyOption('Select employee')}
                  {activeEmployees.map((e) => (
                    <MenuItem key={e.id} value={e.id}>
                      {renderEmployeeLabel(e, { showInactive: false })}
                    </MenuItem>
                  ))}
                </TextField>
              )} />
              <Controller name="service_id" control={control} render={({ field }) => (
                <TextField {...field} select label="Service" fullWidth error={!!errors.service_id} helperText={errors.service_id?.message}>
                  {emptyOption('Select service')}
                  {activeServices.map((s) => <MenuItem key={s.id} value={s.id}>{s.name} ({s.duration_minutes} min)</MenuItem>)}
                </TextField>
              )} />
              <Controller name="start_at" control={control} render={({ field }) => (
                <BookingDateTimeFields
                  value={field.value}
                  onChange={field.onChange}
                  dateLabel="Date"
                  timeLabel="Time"
                  error={errors.start_at?.message}
                />
              )} />
              <Controller name="notes" control={control} render={({ field }) => (
                <TextField {...field} label="Notes" multiline rows={2} fullWidth />
              )} />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button type="button" onClick={() => setOpen(false)}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={createMutation.isPending || !canCreateBooking}>
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
    </Box>
  );
}
