import type { SchedulerEvent, SchedulerResource } from '@mui/x-scheduler/models';
import type { Booking, Customer, Employee, Service } from '../../api/client';
import { resolveBookingStatusEventColor } from '../../constants/bookingStatuses';
import { formatPrice } from '../../utils/formatCurrency';
import { bookingClassName } from './schedulerEventUtils';

export interface MeetoriaSchedulerEvent extends SchedulerEvent {
  meetoria?: {
    bookingId: string;
    customerId: string;
    employeeId: string;
    serviceId: string;
    status: string;
    price: number;
    currency: string;
  };
}

export function buildSchedulerResources(employees: Employee[]): SchedulerResource[] {
  return employees
    .filter((employee) => employee.is_active)
    .map((employee) => ({
      id: employee.id,
      title: `${employee.first_name} ${employee.last_name}`,
    }));
}

export function buildSchedulerEvents(
  bookings: Booking[],
  customers: Customer[],
  services: Service[],
): MeetoriaSchedulerEvent[] {
  const customerMap = new Map(customers.map((customer) => [customer.id, customer]));
  const serviceMap = new Map(services.map((service) => [service.id, service]));

  return bookings
    .map((booking) => {
      const customer = customerMap.get(booking.customer_id);
      const service = serviceMap.get(booking.service_id);
      const customerName = customer
        ? `${customer.first_name} ${customer.last_name}`.trim()
        : 'Customer';
      const serviceName = service?.name ?? 'Service';
      const isLocked = booking.status === 'completed'
        || booking.status === 'no_show'
        || booking.status === 'cancelled';

      return {
        id: booking.id,
        title: customerName,
        start: booking.start_time,
        end: booking.end_time,
        resource: booking.employee_id,
        className: bookingClassName(booking.id),
        color: resolveBookingStatusEventColor(booking.status),
        readOnly: isLocked,
        draggable: !isLocked,
        resizable: false,
        description: [
          `Service: ${serviceName}`,
          `Status: ${booking.status.replace('_', ' ')}`,
          `Price: ${formatPrice(booking.price, booking.currency)}`,
          booking.notes ? `Notes: ${booking.notes}` : '',
        ]
          .filter(Boolean)
          .join('\n'),
        meetoria: {
          bookingId: booking.id,
          customerId: booking.customer_id,
          employeeId: booking.employee_id,
          serviceId: booking.service_id,
          status: booking.status,
          price: booking.price,
          currency: booking.currency,
        },
      };
    });
}

export interface BookingScheduleChange {
  bookingId: string;
  startTime: string;
  employeeId: string;
}

export function findScheduleChanges(
  previous: Booking[],
  nextEvents: MeetoriaSchedulerEvent[],
): BookingScheduleChange[] {
  const previousById = new Map(previous.map((booking) => [booking.id, booking]));
  const changes: BookingScheduleChange[] = [];

  for (const event of nextEvents) {
    const bookingId = String(event.id);
    const original = previousById.get(bookingId);
    if (!original) {
      continue;
    }

    const nextEmployeeId = event.resource ? String(event.resource) : original.employee_id;
    const startChanged = original.start_time !== event.start;
    const employeeChanged = original.employee_id !== nextEmployeeId;

    if (startChanged || employeeChanged) {
      changes.push({
        bookingId,
        startTime: event.start,
        employeeId: nextEmployeeId,
      });
    }
  }

  return changes;
}
