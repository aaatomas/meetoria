import { useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useMutation, useQuery } from '@tanstack/react-query';
import dayjs, { type Dayjs } from 'dayjs';
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  CircularProgress,
  Container,
  Divider,
  Stack,
  Step,
  StepLabel,
  Stepper,
  TextField,
  Typography,
} from '@mui/material';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { BookingDateField } from '../components/bookings/BookingDateTimeFields';
import {
  createPublicBooking,
  getPublicAvailability,
  getPublicBranches,
  getPublicEmployees,
  getPublicOrganization,
  getPublicServices,
  getApiErrorMessage,
  type PublicBranch,
  type PublicEmployee,
  type PublicService,
  type PublicTimeSlot,
} from '../api/publicClient';
import { formatPrice } from '../utils/formatCurrency';
import { requiredPhoneField, formatPhoneDisplay } from '../utils/phoneUtils';
import { PhoneField } from '../components/common/PhoneField';

const ANY_EMPLOYEE = '__any__';

const BASE_STEPS = ['Service', 'Staff', 'Date & Time', 'Your Details', 'Confirm'] as const;
const LOCATION_STEP = 'Location';

function formatTime(iso: string, timeFormat: '24h' | '12h' = '24h'): string {
  return dayjs(iso).format(timeFormat === '12h' ? 'h:mm A' : 'HH:mm');
}

function formatDate(iso: string): string {
  return dayjs(iso).format('dddd, D MMMM YYYY');
}

function employeeName(emp: PublicEmployee): string {
  return `${emp.first_name} ${emp.last_name}`.trim();
}

function branchLabel(branch: PublicBranch): string {
  const parts = [branch.name];
  if (branch.city) parts.push(branch.city);
  return parts.join(' · ');
}

export function PublicBookingPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const [activeStep, setActiveStep] = useState(0);
  const [selectedBranchId, setSelectedBranchId] = useState<string | null>(null);
  const [selectedService, setSelectedService] = useState<PublicService | null>(null);
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>(ANY_EMPLOYEE);
  const [selectedDate, setSelectedDate] = useState<Dayjs | null>(null);
  const selectedDateKey = selectedDate?.format('YYYY-MM-DD') ?? '';
  const [selectedSlot, setSelectedSlot] = useState<PublicTimeSlot | null>(null);
  const [bookingComplete, setBookingComplete] = useState(false);

  const orgQuery = useQuery({
    queryKey: ['public-org', slug],
    queryFn: () => getPublicOrganization(slug),
    enabled: !!slug,
  });

  const branchesQuery = useQuery({
    queryKey: ['public-branches', slug],
    queryFn: () => getPublicBranches(slug),
    enabled: !!slug && !!orgQuery.data,
  });

  const branches = branchesQuery.data ?? [];
  const hasMultipleBranches = branches.length > 1;
  const steps = hasMultipleBranches ? [LOCATION_STEP, ...BASE_STEPS] : [...BASE_STEPS];

  useEffect(() => {
    if (branches.length === 1) {
      setSelectedBranchId(branches[0].id);
    }
  }, [branches]);

  const currency = orgQuery.data?.currency?.trim() || 'EUR';
  const timeFormat = orgQuery.data?.time_format === '12h' ? '12h' : '24h';

  const servicesQuery = useQuery({
    queryKey: ['public-services', slug, selectedBranchId],
    queryFn: () => getPublicServices(slug, selectedBranchId!),
    enabled: !!slug && !!selectedBranchId,
  });

  const employeesQuery = useQuery({
    queryKey: ['public-employees', slug, selectedBranchId, selectedService?.id],
    queryFn: () => getPublicEmployees(slug, selectedBranchId!, selectedService!.id),
    enabled: !!slug && !!selectedBranchId && !!selectedService,
  });

  const availabilityQuery = useQuery({
    queryKey: ['public-availability', slug, selectedBranchId, selectedService?.id, selectedEmployeeId, selectedDateKey],
    queryFn: () =>
      getPublicAvailability(slug, {
        branch_id: selectedBranchId!,
        service_id: selectedService!.id,
        date: selectedDateKey,
        employee_id: selectedEmployeeId === ANY_EMPLOYEE ? undefined : selectedEmployeeId,
      }),
    enabled: !!slug && !!selectedBranchId && !!selectedService && !!selectedDateKey && activeStep >= (hasMultipleBranches ? 3 : 2),
  });

  const customerSchema = useMemo(
    () =>
      z.object({
        first_name: z.string().min(1, 'First name is required'),
        last_name: z.string().min(1, 'Last name is required'),
        phone: requiredPhoneField,
        email: orgQuery.data?.email_required
          ? z.string().email('Valid email is required')
          : z.string().email('Invalid email').optional().or(z.literal('')),
      }),
    [orgQuery.data?.email_required],
  );

  type CustomerForm = z.infer<typeof customerSchema>;

  const { control, handleSubmit, getValues } = useForm<CustomerForm>({
    resolver: zodResolver(customerSchema),
    defaultValues: { first_name: '', last_name: '', phone: '', email: '' },
  });

  const bookingMutation = useMutation({
    mutationFn: () => {
      const customer = getValues();
      const employeeId =
        selectedEmployeeId === ANY_EMPLOYEE
          ? selectedSlot?.employee_ids?.[0]
          : selectedEmployeeId;

      return createPublicBooking(slug, {
        branch_id: selectedBranchId!,
        service_id: selectedService!.id,
        employee_id: employeeId,
        start_time: selectedSlot!.start_time,
        customer: {
          first_name: customer.first_name,
          last_name: customer.last_name,
          phone: customer.phone,
          email: customer.email || undefined,
        },
      });
    },
    onSuccess: () => setBookingComplete(true),
  });

  const availableSlots = (availabilityQuery.data ?? []).filter((s) => s.available);

  const selectedEmployee = employeesQuery.data?.find((e) => e.id === selectedEmployeeId);
  const selectedBranch = branches.find((b) => b.id === selectedBranchId);

  const currentStep = steps[activeStep];

  if (orgQuery.isLoading || branchesQuery.isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
        <CircularProgress />
      </Box>
    );
  }

  if (orgQuery.isError || !orgQuery.data) {
    return (
      <Container maxWidth="sm" sx={{ py: 8 }}>
        <Alert severity="error">
          {getApiErrorMessage(orgQuery.error) || 'This booking page is not available.'}
        </Alert>
      </Container>
    );
  }

  if (branchesQuery.isError || branches.length === 0) {
    return (
      <Container maxWidth="sm" sx={{ py: 8 }}>
        <Alert severity="error">
          {getApiErrorMessage(branchesQuery.error) || 'No locations are available for booking.'}
        </Alert>
      </Container>
    );
  }

  const org = orgQuery.data;

  if (bookingComplete) {
    return (
      <Container maxWidth="sm" sx={{ py: 8 }}>
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h5" fontWeight={700} gutterBottom>
              Booking Confirmed
            </Typography>
            <Typography color="text.secondary" mb={2}>
              Your appointment with {org.name} has been booked.
            </Typography>
            {selectedService && selectedSlot && (
              <Typography>
                {selectedService.name} · {formatDate(selectedSlot.start_time)} at{' '}
                {formatTime(selectedSlot.start_time, timeFormat)}
              </Typography>
            )}
          </CardContent>
        </Card>
      </Container>
    );
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default', py: 4 }}>
      <Container maxWidth="md">
        <Stack spacing={1} mb={4} textAlign="center">
          <Typography variant="h4" fontWeight={700}>
            {org.name}
          </Typography>
          <Typography color="text.secondary">Book an appointment online</Typography>
        </Stack>

        <Stepper activeStep={activeStep} alternativeLabel sx={{ mb: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <Card>
          <CardContent sx={{ p: { xs: 2, sm: 4 } }}>
            {currentStep === LOCATION_STEP && (
              <Stack spacing={2}>
                <Typography variant="h6">Select a location</Typography>
                {branches.map((branch) => (
                  <Card
                    key={branch.id}
                    variant="outlined"
                    sx={{ borderColor: selectedBranchId === branch.id ? 'primary.main' : undefined }}
                  >
                    <CardActionArea
                      onClick={() => {
                        setSelectedBranchId(branch.id);
                        setSelectedService(null);
                        setSelectedEmployeeId(ANY_EMPLOYEE);
                        setSelectedDate(null);
                        setSelectedSlot(null);
                      }}
                    >
                      <CardContent>
                        <Typography fontWeight={600}>{branchLabel(branch)}</Typography>
                        {branch.address && (
                          <Typography variant="body2" color="text.secondary">
                            {branch.address}
                          </Typography>
                        )}
                        {branch.phone && (
                          <Typography variant="body2" color="text.secondary" mt={0.5}>
                            {formatPhoneDisplay(branch.phone)}
                          </Typography>
                        )}
                      </CardContent>
                    </CardActionArea>
                  </Card>
                ))}
              </Stack>
            )}

            {currentStep === 'Service' && (
              <Stack spacing={2}>
                <Typography variant="h6">Select a service</Typography>
                {servicesQuery.isLoading && <CircularProgress size={24} />}
                {servicesQuery.data?.map((service) => (
                  <Card
                    key={service.id}
                    variant="outlined"
                    sx={{
                      borderColor: selectedService?.id === service.id ? 'primary.main' : undefined,
                    }}
                  >
                    <CardActionArea
                      onClick={() => {
                        setSelectedService(service);
                        setSelectedEmployeeId(ANY_EMPLOYEE);
                        setSelectedDate(null);
                        setSelectedSlot(null);
                      }}
                    >
                      <CardContent>
                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                          <Box>
                            <Typography fontWeight={600}>{service.name}</Typography>
                            {service.description && (
                              <Typography variant="body2" color="text.secondary">
                                {service.description}
                              </Typography>
                            )}
                            <Typography variant="body2" color="text.secondary" mt={0.5}>
                              {service.duration_minutes} min
                            </Typography>
                          </Box>
                          <Typography fontWeight={600}>
                            {formatPrice(service.price, currency)}
                          </Typography>
                        </Stack>
                      </CardContent>
                    </CardActionArea>
                  </Card>
                ))}
              </Stack>
            )}

            {currentStep === 'Staff' && (
              <Stack spacing={2}>
                <Typography variant="h6">Select staff member</Typography>
                <Card
                  variant="outlined"
                  sx={{ borderColor: selectedEmployeeId === ANY_EMPLOYEE ? 'primary.main' : undefined }}
                >
                  <CardActionArea onClick={() => setSelectedEmployeeId(ANY_EMPLOYEE)}>
                    <CardContent>
                      <Typography fontWeight={600}>Any Available</Typography>
                      <Typography variant="body2" color="text.secondary">
                        We&apos;ll assign the first available staff member
                      </Typography>
                    </CardContent>
                  </CardActionArea>
                </Card>
                {employeesQuery.data?.map((emp) => (
                  <Card
                    key={emp.id}
                    variant="outlined"
                    sx={{ borderColor: selectedEmployeeId === emp.id ? 'primary.main' : undefined }}
                  >
                    <CardActionArea onClick={() => setSelectedEmployeeId(emp.id)}>
                      <CardContent>
                        <Typography fontWeight={600}>{employeeName(emp)}</Typography>
                        {emp.title && (
                          <Typography variant="body2" color="text.secondary">
                            {emp.title}
                          </Typography>
                        )}
                      </CardContent>
                    </CardActionArea>
                  </Card>
                ))}
              </Stack>
            )}

            {currentStep === 'Date & Time' && (
              <Stack spacing={3}>
                <Typography variant="h6">Select date and time</Typography>
                <BookingDateField
                  label="Date"
                  value={selectedDate}
                  minDate={dayjs().startOf('day')}
                  onChange={(date) => {
                    setSelectedDate(date);
                    setSelectedSlot(null);
                  }}
                />
                {selectedDate && availabilityQuery.isError && (
                  <Alert severity="error">{getApiErrorMessage(availabilityQuery.error)}</Alert>
                )}
                {selectedDate && availabilityQuery.isLoading && <CircularProgress size={24} />}
                {selectedDate && !availabilityQuery.isLoading && availableSlots.length === 0 && (
                  <Alert severity="info">
                    No available times on this date.
                    {[0, 6].includes(selectedDate.day())
                      ? ' The business may be closed on weekends — try a weekday.'
                      : ' Try another day, or ask the business to configure working hours in Settings.'}
                  </Alert>
                )}
                {availableSlots.length > 0 && (
                  <Box display="flex" flexWrap="wrap" gap={1}>
                    {availableSlots.map((slot) => (
                      <Button
                        key={slot.start_time}
                        variant={selectedSlot?.start_time === slot.start_time ? 'contained' : 'outlined'}
                        onClick={() => setSelectedSlot(slot)}
                        sx={{ minWidth: 72, width: 72, px: 0 }}
                      >
                        {formatTime(slot.start_time, timeFormat)}
                      </Button>
                    ))}
                  </Box>
                )}
              </Stack>
            )}

            {currentStep === 'Your Details' && (
              <Stack spacing={2} component="form" id="customer-form">
                <Typography variant="h6">Your information</Typography>
                <Controller
                  name="first_name"
                  control={control}
                  render={({ field, fieldState }) => (
                    <TextField {...field} label="First name" required fullWidth error={!!fieldState.error} helperText={fieldState.error?.message} />
                  )}
                />
                <Controller
                  name="last_name"
                  control={control}
                  render={({ field, fieldState }) => (
                    <TextField {...field} label="Last name" required fullWidth error={!!fieldState.error} helperText={fieldState.error?.message} />
                  )}
                />
                <Controller
                  name="phone"
                  control={control}
                  render={({ field, fieldState }) => (
                    <PhoneField
                      {...field}
                      label="Phone"
                      required
                      fullWidth
                      error={!!fieldState.error}
                      helperText={fieldState.error?.message}
                    />
                  )}
                />
                <Controller
                  name="email"
                  control={control}
                  render={({ field, fieldState }) => (
                    <TextField {...field} label="Email" fullWidth required={org.email_required} error={!!fieldState.error} helperText={fieldState.error?.message} />
                  )}
                />
              </Stack>
            )}

            {currentStep === 'Confirm' && selectedService && selectedSlot && (
              <Stack spacing={2}>
                <Typography variant="h6">Confirm your booking</Typography>
                <Divider />
                <Stack spacing={1}>
                  {selectedBranch && <Row label="Location" value={branchLabel(selectedBranch)} />}
                  <Row label="Service" value={selectedService.name} />
                  <Row
                    label="Staff"
                    value={
                      selectedEmployeeId === ANY_EMPLOYEE
                        ? 'Any Available'
                        : selectedEmployee
                          ? employeeName(selectedEmployee)
                          : '—'
                    }
                  />
                  <Row label="Date" value={formatDate(selectedSlot.start_time)} />
                  <Row label="Time" value={formatTime(selectedSlot.start_time, timeFormat)} />
                  <Row
                    label="Price"
                    value={formatPrice(selectedService.price, currency)}
                  />
                </Stack>
                {(org.cancellation_policy || org.rescheduling_policy) && (
                  <>
                    <Divider />
                    {org.cancellation_policy && (
                      <Typography variant="body2" color="text.secondary">
                        <strong>Cancellation:</strong> {org.cancellation_policy}
                      </Typography>
                    )}
                    {org.rescheduling_policy && (
                      <Typography variant="body2" color="text.secondary">
                        <strong>Rescheduling:</strong> {org.rescheduling_policy}
                      </Typography>
                    )}
                  </>
                )}
                {bookingMutation.isError && (
                  <Alert severity="error">{getApiErrorMessage(bookingMutation.error)}</Alert>
                )}
              </Stack>
            )}

            <Stack direction="row" justifyContent="space-between" mt={4}>
              <Button
                disabled={activeStep === 0 || bookingMutation.isPending}
                onClick={() => setActiveStep((s) => s - 1)}
              >
                Back
              </Button>
              {activeStep < steps.length - 1 ? (
                <Button
                  variant="contained"
                  disabled={
                    (currentStep === LOCATION_STEP && !selectedBranchId) ||
                    (currentStep === 'Service' && !selectedService) ||
                    (currentStep === 'Staff' && !selectedEmployeeId) ||
                    (currentStep === 'Date & Time' && !selectedSlot)
                  }
                  onClick={() => {
                    if (currentStep === 'Your Details') {
                      handleSubmit(() => setActiveStep((s) => s + 1))();
                    } else {
                      setActiveStep((s) => s + 1);
                    }
                  }}
                >
                  Next
                </Button>
              ) : (
                <Button
                  variant="contained"
                  disabled={bookingMutation.isPending}
                  onClick={() => bookingMutation.mutate()}
                >
                  {bookingMutation.isPending ? 'Booking…' : 'Confirm Booking'}
                </Button>
              )}
            </Stack>
          </CardContent>
        </Card>
      </Container>
    </Box>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <Stack direction="row" justifyContent="space-between">
      <Typography color="text.secondary">{label}</Typography>
      <Typography fontWeight={500}>{value}</Typography>
    </Stack>
  );
}
