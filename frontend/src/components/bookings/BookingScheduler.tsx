import { useCallback, useMemo, useRef, useState } from 'react';
import { Box, CircularProgress, Typography } from '@mui/material';
import { EventCalendar } from '@mui/x-scheduler/event-calendar';
import type { SchedulerEvent } from '@mui/x-scheduler/models';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, Booking, Customer, Employee, Service, getApiErrorMessage } from '../../api/client';
import {
  buildSchedulerEvents,
  buildSchedulerResources,
  findScheduleChanges,
  MeetoriaSchedulerEvent,
} from './bookingSchedulerUtils';
import { useSchedulerResourceAvatars } from './useSchedulerResourceAvatars';
import {
  buildVisibleResourcesState,
  mergeVisibleResources,
  useSchedulerResourceFilterActions,
} from './useSchedulerResourceFilterActions';
import { useSchedulerEventClick } from './useSchedulerEventClick';
import { BookingEventDialog, EditBookingForm } from './BookingEventDialog';
import { bookingSchedulerLocaleText } from './bookingSchedulerLocale';

const defaultSchedulerPreferences = {
  ampm: false,
  isSidePanelOpen: false,
} as const;

interface BookingSchedulerProps {
  orgId: string;
  bookings: Booking[];
  customers: Customer[];
  employees: Employee[];
  services: Service[];
  isLoading: boolean;
}

export function BookingScheduler({
  orgId,
  bookings,
  customers,
  employees,
  services,
  isLoading,
}: BookingSchedulerProps) {
  const queryClient = useQueryClient();
  const schedulerRef = useRef<HTMLDivElement>(null);
  const bookingsRef = useRef(bookings);
  bookingsRef.current = bookings;

  const [selectedBookingId, setSelectedBookingId] = useState<string | null>(null);
  const [dialogError, setDialogError] = useState<string | null>(null);

  const activeEmployees = useMemo(
    () => employees.filter((employee) => employee.is_active),
    [employees],
  );

  useSchedulerResourceAvatars(schedulerRef, activeEmployees);

  const resources = useMemo(() => buildSchedulerResources(employees), [employees]);
  const resourceIds = useMemo(() => resources.map((resource) => String(resource.id)), [resources]);
  const [visibleResources, setVisibleResources] = useState<Record<string, boolean>>({});

  const resolvedVisibleResources = useMemo(
    () => mergeVisibleResources(resourceIds, visibleResources),
    [resourceIds, visibleResources],
  );

  const handleVisibleResourcesChange = useCallback(
    (nextVisibleResources: Record<string, boolean>) => {
      setVisibleResources(mergeVisibleResources(resourceIds, nextVisibleResources));
    },
    [resourceIds],
  );

  const handleSelectAllEmployees = useCallback(() => {
    setVisibleResources(buildVisibleResourcesState(resourceIds, visibleResources, true));
  }, [resourceIds, visibleResources]);

  const handleDeselectAllEmployees = useCallback(() => {
    setVisibleResources(buildVisibleResourcesState(resourceIds, visibleResources, false));
  }, [resourceIds, visibleResources]);

  useSchedulerResourceFilterActions(
    schedulerRef,
    handleSelectAllEmployees,
    handleDeselectAllEmployees,
  );

  const events = useMemo(
    () => buildSchedulerEvents(bookings, customers, services),
    [bookings, customers, services],
  );

  const selectedBooking = useMemo(
    () => bookings.find((booking) => booking.id === selectedBookingId) ?? null,
    [bookings, selectedBookingId],
  );

  const handleBookingClick = useCallback((bookingId: string) => {
    setDialogError(null);
    setSelectedBookingId(bookingId);
  }, []);

  const isSchedulerReady = !isLoading && employees.length > 0;

  useSchedulerEventClick(schedulerRef, handleBookingClick, isSchedulerReady);

  const closeDialog = useCallback(() => {
    setSelectedBookingId(null);
    setDialogError(null);
  }, []);

  const updateBookingMutation = useMutation({
    mutationFn: async ({
      bookingId,
      data,
    }: {
      bookingId: string;
      data: EditBookingForm;
    }) => {
      await api.put(`/organizations/${orgId}/bookings/${bookingId}`, {
        employee_id: data.employee_id,
        service_id: data.service_id,
        start_time: data.start_at!.toISOString(),
        notes: data.notes ?? '',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
      closeDialog();
    },
    onError: (error) => {
      setDialogError(getApiErrorMessage(error));
    },
  });

  const handleDialogSave = useCallback(
    (bookingId: string, data: EditBookingForm) => {
      setDialogError(null);
      updateBookingMutation.mutate({ bookingId, data });
    },
    [updateBookingMutation],
  );

  const updateScheduleMutation = useMutation({
    mutationFn: async ({
      bookingId,
      startTime,
      employeeId,
    }: {
      bookingId: string;
      startTime: string;
      employeeId: string;
    }) => {
      const original = bookingsRef.current.find((booking) => booking.id === bookingId);
      await api.put(`/organizations/${orgId}/bookings/${bookingId}`, {
        start_time: startTime,
        ...(original && original.employee_id !== employeeId
          ? { employee_id: employeeId }
          : {}),
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings'] });
    },
  });

  const handleEventsChange = useCallback(
    (nextEvents: SchedulerEvent[]) => {
      const typedEvents = nextEvents as MeetoriaSchedulerEvent[];
      const changes = findScheduleChanges(bookingsRef.current, typedEvents);
      const pendingChange = changes[0];

      if (pendingChange && !updateScheduleMutation.isPending) {
        updateScheduleMutation.mutate({
          bookingId: pendingChange.bookingId,
          startTime: pendingChange.startTime,
          employeeId: pendingChange.employeeId,
        });
      }
    },
    [updateScheduleMutation],
  );

  return (
    <>
      <Box
        ref={schedulerRef}
        sx={{
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        {isLoading ? (
          <Box sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <CircularProgress />
          </Box>
        ) : employees.length === 0 ? (
          <Box sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 3 }}>
            <Typography color="text.secondary">
              Add employees to display the booking schedule.
            </Typography>
          </Box>
        ) : (
          <EventCalendar
            events={events}
            resources={resources}
            visibleResources={resolvedVisibleResources}
            onVisibleResourcesChange={handleVisibleResourcesChange}
            onEventsChange={handleEventsChange}
            localeText={bookingSchedulerLocaleText}
            defaultPreferences={defaultSchedulerPreferences}
            views={['week', 'day', 'month', 'agenda']}
            defaultView="week"
            viewConfig={{
              week: { startTime: 8, endTime: 20 },
              day: { startTime: 8, endTime: 20 },
            }}
            eventCreation={false}
            areEventsDraggable
            areEventsResizable={false}
            shouldEventRequireResource
            sx={{ flex: 1, minHeight: 0, height: '100%' }}
          />
        )}
      </Box>

      <BookingEventDialog
        orgId={orgId}
        open={selectedBookingId !== null}
        booking={selectedBooking}
        customers={customers}
        employees={employees}
        services={services}
        isSaving={updateBookingMutation.isPending}
        errorMessage={dialogError}
        onClose={closeDialog}
        onSave={handleDialogSave}
      />
    </>
  );
}
