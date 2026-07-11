import type { EventCalendarLocaleText } from '@mui/x-scheduler/models';

export const bookingSchedulerLocaleText: Partial<EventCalendarLocaleText> = {
  resourcesLabel: 'Employees',
  resourceLabel: 'Employee',
  resourceAriaLabel: (employeeName) => `Employee: ${employeeName}`,
  resourceColorSectionLabel: 'Employees & status',
  colorPickerLabel: 'Status',
  labelNoResource: 'No employee',
  labelInvalidResource: 'Invalid employee',
  noResourceAriaLabel: 'No specific employee',
  requiredResourceError: 'An employee is required.',
  descriptionLabel: 'Details',
  timelineResourceTitleHeader: 'Employee',
};
