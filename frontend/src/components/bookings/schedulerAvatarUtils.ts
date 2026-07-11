import type { Employee } from '../../api/client';

export function employeeDecorationKey(employee: Employee): string {
  return `${employee.id}:${employee.updated_at ?? ''}:${employee.avatar_url ?? ''}`;
}
