import axios from 'axios';
import { getToken } from '../auth/keycloak';

const API_URL = import.meta.env.VITE_API_URL || '';

export function getApiErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as { message?: string } | undefined;
    if (data?.message) return data.message;
    if (error.message) return error.message;
  }
  if (error instanceof Error) return error.message;
  return 'Something went wrong. Please try again.';
}

export const api = axios.create({
  baseURL: `${API_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  const orgId = localStorage.getItem('organization_id');
  if (orgId) {
    config.headers['X-Organization-ID'] = orgId;
  }

  const branchId = localStorage.getItem('branch_id');
  if (branchId) {
    config.headers['X-Branch-ID'] = branchId;
  }

  if (config.data instanceof FormData) {
    delete config.headers['Content-Type'];
  }

  return config;
});

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  timezone: string;
  currency: string;
  email?: string;
  phone?: string;
  is_active: boolean;
  settings?: string;
}

export interface Branch {
  id: string;
  organization_id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  timezone?: string;
  phone?: string;
  email?: string;
  is_active: boolean;
  is_default: boolean;
}

export function getActiveBranchId(): string | null {
  return localStorage.getItem('branch_id');
}

export function pickActiveBranch(branches: Branch[], preferredId?: string | null): Branch | undefined {
  const active = branches.filter((branch) => branch.is_active);
  if (preferredId) {
    const preferred = active.find((branch) => branch.id === preferredId);
    if (preferred) return preferred;
  }
  return active.find((branch) => branch.is_default) ?? active[0];
}

export async function resolveActiveBranchId(orgId: string): Promise<string | null> {
  const branches = await listBranches(orgId);
  const branch = pickActiveBranch(branches, getActiveBranchId());
  if (branch) {
    setActiveBranchId(branch.id);
    return branch.id;
  }
  return null;
}

export function setActiveOrganizationId(orgId: string): void {
  localStorage.setItem('organization_id', orgId);
  localStorage.removeItem('branch_id');
}

export function setActiveBranchId(branchId: string): void {
  localStorage.setItem('branch_id', branchId);
}

export function locationKey(orgId: string, branchId: string): string {
  return `${orgId}:${branchId}`;
}

export function parseLocationKey(key: string): { orgId: string; branchId: string } | null {
  const [orgId, branchId] = key.split(':');
  if (!orgId || !branchId) return null;
  return { orgId, branchId };
}

export function setActiveLocation(orgId: string, branchId: string): void {
  localStorage.setItem('organization_id', orgId);
  localStorage.setItem('branch_id', branchId);
}

export async function listBranches(orgId: string): Promise<Branch[]> {
  const { data } = await api.get<PaginatedResponse<Branch>>(`/organizations/${orgId}/branches`, {
    params: { limit: 100 },
  });
  return data.data;
}

export async function createBranch(
  orgId: string,
  payload: Pick<Branch, 'name'> & Partial<Pick<Branch, 'address' | 'city' | 'country' | 'timezone' | 'phone' | 'email'>>,
): Promise<Branch> {
  const { data } = await api.post<Branch>(`/organizations/${orgId}/branches`, payload);
  return data;
}

export async function updateBranch(
  orgId: string,
  branchId: string,
  payload: Partial<Pick<Branch, 'name' | 'address' | 'city' | 'country' | 'timezone' | 'phone' | 'email' | 'is_active'>>,
): Promise<Branch> {
  const { data } = await api.put<Branch>(`/organizations/${orgId}/branches/${branchId}`, payload);
  return data;
}

export async function deleteBranch(orgId: string, branchId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/branches/${branchId}`);
}

export async function setDefaultBranch(orgId: string, branchId: string): Promise<Branch> {
  const { data } = await api.post<Branch>(`/organizations/${orgId}/branches/${branchId}/set-default`);
  return data;
}

export interface BookingSettings {
  enabled: boolean;
  booking_window_days: number;
  min_notice_minutes: number;
  max_notice_minutes?: number;
  email_required: boolean;
  auto_confirm: boolean;
  manual_approval: boolean;
  cancellation_policy?: string;
  rescheduling_policy?: string;
}

export function parseOrganizationSettings(settings?: string): {
  booking: BookingSettings;
  time_format: '24h' | '12h';
} {
  const defaults: BookingSettings = {
    enabled: true,
    booking_window_days: 30,
    min_notice_minutes: 60,
    email_required: false,
    auto_confirm: true,
    manual_approval: false,
  };
  if (!settings || settings === '{}') {
    return { booking: defaults, time_format: '24h' };
  }
  try {
    const parsed = JSON.parse(settings) as {
      booking?: Partial<BookingSettings>;
      time_format?: string;
    };
    const booking = { ...defaults, ...parsed.booking };
    if (booking.max_notice_minutes != null && booking.max_notice_minutes <= 0) {
      delete booking.max_notice_minutes;
    }
    const time_format = parsed.time_format === '12h' ? '12h' : '24h';
    return { booking, time_format };
  } catch {
    return { booking: defaults, time_format: '24h' };
  }
}

export interface DaySchedule {
  day_of_week: number;
  slots: Array<{ start_time: string; end_time: string }>;
}

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export function dayName(dayOfWeek: number): string {
  return DAY_NAMES[dayOfWeek] ?? `Day ${dayOfWeek}`;
}

export function defaultWeekSchedule(): DaySchedule[] {
  return Array.from({ length: 7 }, (_, day) => ({
    day_of_week: day,
    slots: day >= 1 && day <= 5 ? [{ start_time: '09:00', end_time: '17:00' }] : [],
  }));
}

export async function getWorkingHours(orgId: string, branchId?: string): Promise<DaySchedule[]> {
  const params = branchId ? { branch_id: branchId } : undefined;
  const { data } = await api.get<{ schedule: DaySchedule[] }>(`/organizations/${orgId}/schedule/working-hours`, {
    params,
  });
  const schedule = data.schedule ?? [];
  if (schedule.length === 0) return defaultWeekSchedule();
  const byDay = new Map(schedule.map((d) => [d.day_of_week, d]));
  return Array.from({ length: 7 }, (_, day) => byDay.get(day) ?? { day_of_week: day, slots: [] });
}

export async function saveWorkingHours(orgId: string, schedule: DaySchedule[], branchId?: string): Promise<DaySchedule[]> {
  const { data } = await api.put<{ schedule: DaySchedule[] }>(`/organizations/${orgId}/schedule/working-hours`, {
    schedule,
    ...(branchId ? { branch_id: branchId } : {}),
  });
  return data.schedule;
}

export interface Customer {
  id: string;
  organization_id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  notes?: string;
  bookings_count?: number;
  cancellations_count?: number;
}

export interface Employee {
  id: string;
  organization_id: string;
  branch_id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  title?: string;
  bio?: string;
  avatar_url?: string;
  is_active: boolean;
  color?: string;
  updated_at?: string;
}

export function resolveUploadUrl(url?: string): string | undefined {
  if (!url) return undefined;
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) {
    return url;
  }
  return url;
}

export async function uploadEmployeeAvatar(orgId: string, employeeId: string, file: File): Promise<Employee> {
  const formData = new FormData();
  formData.append('avatar', file);
  const { data } = await api.post<Employee>(`/organizations/${orgId}/employees/${employeeId}/avatar`, formData);
  return data;
}

export interface DeletionCheck {
  can_delete: boolean;
  bookings_count: number;
  employees_count?: number;
  message?: string;
}

export async function listEmployees(orgId: string, branchId?: string | null): Promise<Employee[]> {
  const params: Record<string, string | number> = { limit: 1000 };
  if (branchId) params.branch_id = branchId;
  const { data } = await api.get<PaginatedResponse<Employee>>(`/organizations/${orgId}/employees`, { params });
  return data.data;
}

export async function listServices(orgId: string, branchId?: string | null): Promise<Service[]> {
  const params: Record<string, string | number> = { limit: 1000 };
  if (branchId) params.branch_id = branchId;
  const { data } = await api.get<PaginatedResponse<Service>>(`/organizations/${orgId}/services`, { params });
  return data.data;
}

export async function checkBranchDeletion(orgId: string, branchId: string): Promise<DeletionCheck> {
  const { data } = await api.get<DeletionCheck>(`/organizations/${orgId}/branches/${branchId}/deletion-check`);
  return data;
}

export async function checkServiceDeletion(orgId: string, serviceId: string): Promise<DeletionCheck> {
  const { data } = await api.get<DeletionCheck>(`/organizations/${orgId}/services/${serviceId}/deletion-check`);
  return data;
}

export async function checkEmployeeDeletion(orgId: string, employeeId: string): Promise<DeletionCheck> {
  const { data } = await api.get<DeletionCheck>(`/organizations/${orgId}/employees/${employeeId}/deletion-check`);
  return data;
}

export async function checkCustomerDeletion(orgId: string, customerId: string): Promise<DeletionCheck> {
  const { data } = await api.get<DeletionCheck>(`/organizations/${orgId}/customers/${customerId}/deletion-check`);
  return data;
}

export async function updateService(
  orgId: string,
  serviceId: string,
  data: Partial<Pick<Service, 'name' | 'description' | 'duration_minutes' | 'price' | 'is_active'>>,
): Promise<Service> {
  const { data: result } = await api.put<Service>(`/organizations/${orgId}/services/${serviceId}`, data);
  return result;
}

export async function deleteService(orgId: string, serviceId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/services/${serviceId}`);
}

export async function updateEmployee(
  orgId: string,
  employeeId: string,
  data: Partial<Pick<Employee, 'first_name' | 'last_name' | 'email' | 'phone' | 'title' | 'bio' | 'is_active'>>,
): Promise<Employee> {
  const { data: result } = await api.put<Employee>(`/organizations/${orgId}/employees/${employeeId}`, data);
  return result;
}

export async function deleteEmployee(orgId: string, employeeId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/employees/${employeeId}`);
}

export async function updateCustomer(
  orgId: string,
  customerId: string,
  data: Partial<Pick<Customer, 'first_name' | 'last_name' | 'email' | 'phone' | 'notes'>>,
): Promise<Customer> {
  const { data: result } = await api.put<Customer>(`/organizations/${orgId}/customers/${customerId}`, data);
  return result;
}

export async function deleteCustomer(orgId: string, customerId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/customers/${customerId}`);
}

export async function sendCustomerSms(orgId: string, customerId: string): Promise<void> {
  await api.post(`/organizations/${orgId}/customers/${customerId}/notifications/sms`);
}

export async function sendCustomerEmail(orgId: string, customerId: string): Promise<void> {
  await api.post(`/organizations/${orgId}/customers/${customerId}/notifications/email`);
}

export async function sendBookingSms(orgId: string, bookingId: string): Promise<void> {
  await api.post(`/organizations/${orgId}/bookings/${bookingId}/notifications/sms`);
}

export async function sendBookingEmail(orgId: string, bookingId: string): Promise<void> {
  await api.post(`/organizations/${orgId}/bookings/${bookingId}/notifications/email`);
}

export interface Notification {
  id: string;
  organization_id: string;
  booking_id?: string;
  channel: 'sms' | 'email';
  template: string;
  recipient: string;
  status: 'created' | 'queued' | 'sent' | 'delivered' | 'failed';
  scheduled_at?: string;
  sent_at?: string;
  delivered_at?: string;
  created_at: string;
  updated_at: string;
}

export async function fetchBookingNotifications(orgId: string, bookingId: string): Promise<Notification[]> {
  const { data } = await api.get<Notification[]>(`/organizations/${orgId}/bookings/${bookingId}/notifications`);
  return data;
}

export async function cancelBooking(orgId: string, bookingId: string, reason: string): Promise<Booking> {
  const { data } = await api.post<Booking>(`/organizations/${orgId}/bookings/${bookingId}/cancel`, { reason });
  return data;
}

export async function updateBookingStatus(orgId: string, bookingId: string, status: string): Promise<Booking> {
  const { data } = await api.put<Booking>(`/organizations/${orgId}/bookings/${bookingId}`, { status });
  return data;
}

export interface Service {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  duration_minutes: number;
  price: number;
  currency: string;
  category?: string;
  color?: string;
  is_active: boolean;
}

export interface Booking {
  id: string;
  organization_id: string;
  branch_id: string;
  customer_id: string;
  employee_id: string;
  service_id: string;
  start_time: string;
  end_time: string;
  status: string;
  price: number;
  currency: string;
  notes?: string;
  cancellation_reason?: string;
}

export type DashboardScope = 'organization' | 'branch';

export interface MetricTrend {
  previous: number;
  change: number;
  change_pct: number | null;
}

export interface DashboardTrends {
  total_bookings: MetricTrend;
  completed_bookings: MetricTrend;
  revenue: MetricTrend;
  new_customers: MetricTrend;
}

export interface DashboardStats {
  scope: 'organization' | 'branch';
  branch_id?: string;
  branch_name?: string;
  total_bookings: number;
  completed_bookings: number;
  cancelled_bookings: number;
  no_show_bookings: number;
  revenue: number;
  new_customers: number;
  trends: DashboardTrends;
  popular_services: Array<{ service_id: string; service_name: string; branch_id: string; branch_name: string; color?: string; count: number; revenue: number }>;
  busiest_days: Array<{ day: string; count: number }>;
  busiest_hours: Array<{ hour: number; count: number }>;
  hourly_heatmap?: HeatmapCell[][];
}

export interface HeatmapCell {
  count: number;
}

export interface DayCount {
  day: string;
  count: number;
}

export interface PopularService {
  service_id: string;
  service_name: string;
  branch_id: string;
  branch_name: string;
  color?: string;
  count: number;
  revenue: number;
}
