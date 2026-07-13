import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  ListItemIcon,
  Menu,
  MenuItem,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material';
import EmailOutlinedIcon from '@mui/icons-material/EmailOutlined';
import CheckIcon from '@mui/icons-material/Check';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import SmsOutlinedIcon from '@mui/icons-material/SmsOutlined';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import dayjs, { Dayjs } from 'dayjs';
import { type ReactNode, useEffect, useRef, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { Booking, Customer, Employee, Service } from '../../api/client';
import { getApiErrorMessage, cancelBooking, fetchBookingNotifications, sendBookingEmail, sendBookingSms, updateBookingStatus } from '../../api/client';
import { BookingNotificationHistory } from './BookingNotificationHistory';
import {
  BOOKING_STATUS_CHANGE_OPTIONS,
  canChangeBookingStatus,
  formatBookingStatus,
  getBookingStatusBadgeColors,
} from '../../constants/bookingStatuses';
import { formatPrice } from '../../utils/formatCurrency';
import { BookingDateTimeFields } from './BookingDateTimeFields';
import { renderEmployeeLabel } from './renderEmployeeLabel';

const editBookingSchema = z.object({
  employee_id: z.string().min(1, 'Employee is required'),
  service_id: z.string().min(1, 'Service is required'),
  start_at: z.custom<Dayjs | null>(
    (value) => value !== null && dayjs.isDayjs(value) && value.isValid(),
    'Start date and time are required',
  ),
  notes: z.string().optional(),
});

export type EditBookingForm = z.infer<typeof editBookingSchema>;

interface BookingEventDialogProps {
  orgId: string;
  currency: string;
  open: boolean;
  booking: Booking | null;
  customers: Customer[];
  employees: Employee[];
  services: Service[];
  isSaving: boolean;
  errorMessage: string | null;
  onClose: () => void;
  onSave: (bookingId: string, data: EditBookingForm) => void;
}

function renderServiceLabel(service: Service) {
  return <span>{service.name}</span>;
}

function buildDialogEmployees(employees: Employee[], booking: Booking | null): Employee[] {
  if (!booking) {
    return employees;
  }

  const activeEmployees = employees.filter((employee) => employee.is_active);
  if (activeEmployees.some((employee) => employee.id === booking.employee_id)) {
    return activeEmployees;
  }

  const currentEmployee = employees.find((employee) => employee.id === booking.employee_id);
  return currentEmployee ? [...activeEmployees, currentEmployee] : activeEmployees;
}

function buildDialogServices(services: Service[], booking: Booking | null): Service[] {
  if (!booking) {
    return services;
  }

  const activeServices = services.filter((service) => service.is_active);
  if (activeServices.some((service) => service.id === booking.service_id)) {
    return activeServices;
  }

  const currentService = services.find((service) => service.id === booking.service_id);
  return currentService ? [...activeServices, currentService] : activeServices;
}

function isBookingLocked(status: string): boolean {
  return status === 'completed' || status === 'no_show' || status === 'cancelled';
}

function TabPanel({ children, value, index }: { children: ReactNode; value: number; index: number }) {
  if (value !== index) {
    return null;
  }

  return <Box sx={{ pt: 2 }}>{children}</Box>;
}

export function BookingEventDialog({
  orgId,
  currency,
  open,
  booking,
  customers,
  employees,
  services,
  isSaving,
  errorMessage,
  onClose,
  onSave,
}: BookingEventDialogProps) {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState(0);
  const [currentStatus, setCurrentStatus] = useState('');
  const locked = isBookingLocked(currentStatus);
  const statusEditable = canChangeBookingStatus(currentStatus);
  const [notificationMessage, setNotificationMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null);
  const [customerMenuAnchor, setCustomerMenuAnchor] = useState<null | HTMLElement>(null);
  const [statusMenuAnchor, setStatusMenuAnchor] = useState<null | HTMLElement>(null);
  const [cancelReasonDialogOpen, setCancelReasonDialogOpen] = useState(false);
  const [cancellationReason, setCancellationReason] = useState('');
  const [cancellationReasonError, setCancellationReasonError] = useState<string | null>(null);
  const [cancellationReasonDisplay, setCancellationReasonDisplay] = useState<string | null>(null);
  const customerMenuOpen = Boolean(customerMenuAnchor);
  const statusMenuOpen = Boolean(statusMenuAnchor);

  const updateStatusMutation = useMutation({
    mutationFn: (status: string) => updateBookingStatus(orgId, booking!.id, status),
    onSuccess: (_, status) => {
      setCurrentStatus(status);
      setStatusMenuAnchor(null);
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
      setStatusMenuAnchor(null);
    },
  });

  const cancelBookingMutation = useMutation({
    mutationFn: (reason: string) => cancelBooking(orgId, booking!.id, reason),
    onSuccess: (_, reason) => {
      setCurrentStatus('cancelled');
      setCancellationReasonDisplay(reason);
      setCancelReasonDialogOpen(false);
      setCancellationReason('');
      setCancellationReasonError(null);
      setStatusMenuAnchor(null);
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const sendSmsMutation = useMutation({
    mutationFn: () => sendBookingSms(orgId, booking!.id),
    onSuccess: () => {
      setNotificationMessage({ severity: 'success', text: 'SMS queued for delivery.' });
      queryClient.invalidateQueries({ queryKey: ['booking-notifications', orgId, booking!.id] });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const sendEmailMutation = useMutation({
    mutationFn: () => sendBookingEmail(orgId, booking!.id),
    onSuccess: () => {
      setNotificationMessage({ severity: 'success', text: 'Email queued for delivery.' });
      queryClient.invalidateQueries({ queryKey: ['booking-notifications', orgId, booking!.id] });
    },
    onError: (error) => {
      setNotificationMessage({ severity: 'error', text: getApiErrorMessage(error) });
    },
  });

  const dialogEmployees = buildDialogEmployees(employees, booking);
  const dialogServices = buildDialogServices(services, booking);

  const {
    control,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<EditBookingForm>({
    resolver: zodResolver(editBookingSchema),
    defaultValues: {
      employee_id: '',
      service_id: '',
      start_at: null,
      notes: '',
    },
  });

  const watchedServiceId = watch('service_id');
  const watchedStartAt = watch('start_at');
  const selectedCustomer = booking
    ? customers.find((item) => item.id === booking.customer_id) ?? null
    : null;
  const selectedService = dialogServices.find((service) => service.id === watchedServiceId);
  const customerName = selectedCustomer
    ? `${selectedCustomer.first_name} ${selectedCustomer.last_name}`.trim()
    : 'Customer';
  const displayPrice = selectedService
    ? { currency, price: selectedService.price }
    : booking
      ? { currency: booking.currency, price: booking.price }
      : null;
  const previewEnd = watchedStartAt && selectedService
    ? watchedStartAt.add(selectedService.duration_minutes, 'minute')
    : booking
      ? dayjs(booking.end_time)
      : null;
  const statusBadgeColors = getBookingStatusBadgeColors(currentStatus || booking?.status || 'pending');

  const notificationsQuery = useQuery({
    queryKey: ['booking-notifications', orgId, booking?.id],
    queryFn: () => fetchBookingNotifications(orgId, booking!.id),
    enabled: open && !!booking,
    refetchInterval: tab === 1 && open ? 3000 : false,
  });

  const initializedBookingIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (!open) {
      initializedBookingIdRef.current = null;
      setTab(0);
      setNotificationMessage(null);
      setCustomerMenuAnchor(null);
      setStatusMenuAnchor(null);
      setCancelReasonDialogOpen(false);
      setCancellationReason('');
      setCancellationReasonError(null);
      setCancellationReasonDisplay(null);
      setCurrentStatus('');
      return;
    }

    if (!booking) {
      return;
    }

    setCurrentStatus(booking.status);
    setCancellationReasonDisplay(booking.cancellation_reason ?? null);

    if (initializedBookingIdRef.current === booking.id) {
      return;
    }

    initializedBookingIdRef.current = booking.id;
    reset({
      employee_id: booking.employee_id,
      service_id: booking.service_id,
      start_at: dayjs(booking.start_time),
      notes: booking.notes ?? '',
    });
  }, [open, booking, reset]);

  const onSubmit = handleSubmit((data) => {
    if (!booking) {
      return;
    }

    onSave(booking.id, data);
  });

  const handleConfirmCancellation = () => {
    const reason = cancellationReason.trim();
    if (!reason) {
      setCancellationReasonError('Cancellation reason is required');
      return;
    }

    setCancellationReasonError(null);
    setNotificationMessage(null);
    cancelBookingMutation.mutate(reason);
  };

  return (
    <>
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ pr: 3 }}>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 2,
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
            <Typography component="span" variant="h6" fontWeight={600} noWrap>
              {customerName}
            </Typography>
            {booking && selectedCustomer && (
              <>
                <IconButton
                  size="small"
                  aria-label="Customer actions"
                  onClick={(event) => setCustomerMenuAnchor(event.currentTarget)}
                  sx={{ flexShrink: 0 }}
                >
                  <MoreVertIcon fontSize="small" />
                </IconButton>
                <Menu
                  anchorEl={customerMenuAnchor}
                  open={customerMenuOpen}
                  onClose={() => setCustomerMenuAnchor(null)}
                  disablePortal
                  anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
                  transformOrigin={{ vertical: 'top', horizontal: 'left' }}
                >
                  <MenuItem
                    disabled={!selectedCustomer.phone || sendSmsMutation.isPending}
                    onClick={() => {
                      setCustomerMenuAnchor(null);
                      setNotificationMessage(null);
                      sendSmsMutation.mutate();
                    }}
                  >
                    <ListItemIcon>
                      <SmsOutlinedIcon fontSize="small" />
                    </ListItemIcon>
                    {sendSmsMutation.isPending ? 'Sending SMS...' : 'Send SMS'}
                  </MenuItem>
                  <MenuItem
                    disabled={!selectedCustomer.email || sendEmailMutation.isPending}
                    onClick={() => {
                      setCustomerMenuAnchor(null);
                      setNotificationMessage(null);
                      sendEmailMutation.mutate();
                    }}
                  >
                    <ListItemIcon>
                      <EmailOutlinedIcon fontSize="small" />
                    </ListItemIcon>
                    {sendEmailMutation.isPending ? 'Sending email...' : 'Send email'}
                  </MenuItem>
                </Menu>
              </>
            )}
          </Box>
          {booking && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexShrink: 0 }}>
              <Typography variant="body2" color="text.secondary" fontWeight={600} noWrap>
                {displayPrice ? formatPrice(displayPrice.price, displayPrice.currency) : ''}
              </Typography>
              <Chip
                size="small"
                label={(
                  <Box sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.25 }}>
                    <span>{formatBookingStatus(currentStatus || booking.status)}</span>
                    {statusEditable && (
                      <KeyboardArrowDownIcon sx={{ fontSize: 16, color: 'inherit' }} />
                    )}
                  </Box>
                )}
                onClick={statusEditable ? (event) => setStatusMenuAnchor(event.currentTarget) : undefined}
                disabled={updateStatusMutation.isPending || cancelBookingMutation.isPending}
                sx={{
                  bgcolor: statusBadgeColors.background,
                  color: statusBadgeColors.foreground,
                  fontWeight: 600,
                  textTransform: 'capitalize',
                  '& .MuiChip-label': {
                    px: 1,
                  },
                  ...(statusEditable && {
                    cursor: 'pointer',
                    '&:hover': {
                      bgcolor: statusBadgeColors.hoverBackground,
                    },
                  }),
                }}
              />
              <Menu
                anchorEl={statusMenuAnchor}
                open={statusMenuOpen}
                onClose={() => setStatusMenuAnchor(null)}
                disablePortal
                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
                transformOrigin={{ vertical: 'top', horizontal: 'right' }}
              >
                {BOOKING_STATUS_CHANGE_OPTIONS.map((status) => (
                  <MenuItem
                    key={status}
                    selected={status === currentStatus}
                    disabled={updateStatusMutation.isPending || cancelBookingMutation.isPending}
                    onClick={() => {
                      if (status === currentStatus || !booking) {
                        setStatusMenuAnchor(null);
                        return;
                      }
                      setNotificationMessage(null);
                      if (status === 'cancelled') {
                        setStatusMenuAnchor(null);
                        setCancellationReason('');
                        setCancellationReasonError(null);
                        setCancelReasonDialogOpen(true);
                        return;
                      }
                      updateStatusMutation.mutate(status);
                    }}
                  >
                    <ListItemIcon sx={{ minWidth: 32 }}>
                      {status === currentStatus ? <CheckIcon fontSize="small" /> : null}
                    </ListItemIcon>
                    <Typography sx={{ textTransform: 'capitalize' }}>
                      {formatBookingStatus(status)}
                    </Typography>
                  </MenuItem>
                ))}
              </Menu>
            </Box>
          )}
        </Box>
      </DialogTitle>
      <Box component="form" onSubmit={onSubmit}>
        <DialogContent>
          <Tabs value={tab} onChange={(_, value) => setTab(value)}>
            <Tab label="Details" />
            <Tab label="Notifications" />
          </Tabs>

          <TabPanel value={tab} index={0}>
          <Stack spacing={2}>
            {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
            {notificationMessage && (
              <Alert severity={notificationMessage.severity}>{notificationMessage.text}</Alert>
            )}

            {currentStatus === 'cancelled' && cancellationReasonDisplay && (
              <Alert severity="warning">
                Cancellation reason: {cancellationReasonDisplay}
              </Alert>
            )}

            <Controller
              name="employee_id"
              control={control}
              render={({ field }) => (
                <TextField
                  select
                  label="Employee"
                  fullWidth
                  disabled={locked}
                  error={!!errors.employee_id}
                  helperText={errors.employee_id?.message}
                  value={field.value ?? ''}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                  inputRef={field.ref}
                  slotProps={{
                    select: {
                      MenuProps: {
                        disablePortal: true,
                      },
                      renderValue: (value) => {
                        const employee = dialogEmployees.find((item) => item.id === value);
                        return employee ? renderEmployeeLabel(employee, { avatarSize: 20 }) : '';
                      },
                    },
                  }}
                >
                  {dialogEmployees.map((employee) => (
                    <MenuItem key={employee.id} value={employee.id}>
                      {renderEmployeeLabel(employee)}
                    </MenuItem>
                  ))}
                </TextField>
              )}
            />

            <Controller
              name="service_id"
              control={control}
              render={({ field }) => (
                <TextField
                  select
                  label="Service"
                  fullWidth
                  disabled={locked}
                  error={!!errors.service_id}
                  helperText={errors.service_id?.message}
                  value={field.value ?? ''}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  name={field.name}
                  inputRef={field.ref}
                  slotProps={{
                    select: {
                      MenuProps: {
                        disablePortal: true,
                      },
                    },
                  }}
                >
                  {dialogServices.map((service) => (
                    <MenuItem key={service.id} value={service.id}>
                      <Box display="flex" alignItems="center" gap={1} width="100%">
                        {renderServiceLabel(service)}
                        {!service.is_active && (
                          <Typography component="span" variant="body2" color="text.secondary">
                            (inactive)
                          </Typography>
                        )}
                      </Box>
                    </MenuItem>
                  ))}
                </TextField>
              )}
            />

            <Controller
              name="start_at"
              control={control}
              render={({ field }) => (
                <BookingDateTimeFields
                  value={field.value}
                  onChange={field.onChange}
                  dateLabel="Start date"
                  timeLabel="Start time"
                  disabled={locked}
                  error={errors.start_at?.message}
                />
              )}
            />

            <TextField
              label="End time"
              value={previewEnd ? previewEnd.format('D MMM YYYY, HH:mm') : ''}
              fullWidth
              disabled
              size="small"
            />

            <Controller
              name="notes"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  label="Notes"
                  multiline
                  rows={4}
                  fullWidth
                  disabled={locked}
                />
              )}
            />
          </Stack>
          </TabPanel>

          <TabPanel value={tab} index={1}>
            <BookingNotificationHistory
              notifications={notificationsQuery.data ?? []}
              isLoading={notificationsQuery.isLoading}
              errorMessage={notificationsQuery.error ? getApiErrorMessage(notificationsQuery.error) : null}
            />
          </TabPanel>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2, gap: 1 }}>
          <Button onClick={onClose}>Close</Button>
          {tab === 0 && !locked && (
            <Button type="submit" variant="contained" disabled={isSaving || updateStatusMutation.isPending || cancelBookingMutation.isPending}>
              {isSaving ? 'Saving...' : 'Save'}
            </Button>
          )}
        </DialogActions>
      </Box>
    </Dialog>

    <Dialog
      open={cancelReasonDialogOpen}
      onClose={() => {
        if (cancelBookingMutation.isPending) {
          return;
        }
        setCancelReasonDialogOpen(false);
        setCancellationReason('');
        setCancellationReasonError(null);
      }}
      maxWidth="xs"
      fullWidth
    >
      <DialogTitle>Cancel booking</DialogTitle>
      <DialogContent>
        <TextField
          autoFocus
          label="Cancellation reason"
          value={cancellationReason}
          onChange={(event) => {
            setCancellationReason(event.target.value);
            if (cancellationReasonError) {
              setCancellationReasonError(null);
            }
          }}
          error={!!cancellationReasonError}
          helperText={cancellationReasonError ?? 'Required'}
          multiline
          rows={3}
          fullWidth
          required
          disabled={cancelBookingMutation.isPending}
          sx={{ mt: 1 }}
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button
          onClick={() => {
            setCancelReasonDialogOpen(false);
            setCancellationReason('');
            setCancellationReasonError(null);
          }}
          disabled={cancelBookingMutation.isPending}
        >
          Back
        </Button>
        <Button
          variant="contained"
          color="error"
          onClick={handleConfirmCancellation}
          disabled={cancelBookingMutation.isPending}
        >
          {cancelBookingMutation.isPending ? 'Cancelling...' : 'Cancel booking'}
        </Button>
      </DialogActions>
    </Dialog>
    </>
  );
}
