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
  email?: string;
  phone?: string;
  is_active: boolean;
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

export interface DashboardStats {
  total_bookings: number;
  completed_bookings: number;
  cancelled_bookings: number;
  no_show_bookings: number;
  revenue: number;
  new_customers: number;
  popular_services: Array<{ service_id: string; service_name: string; color?: string; count: number; revenue: number }>;
  busiest_days: Array<{ day: string; count: number }>;
  busiest_hours: Array<{ hour: number; count: number }>;
  hourly_heatmap?: HeatmapCell[][];
}

export interface HeatmapCell {
  count: number;
}

export interface PopularService {
  service_id: string;
  service_name: string;
  color?: string;
  count: number;
  revenue: number;
}
