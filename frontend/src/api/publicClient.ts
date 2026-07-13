import axios from 'axios';
import { getApiErrorMessage } from './client';

const API_URL = import.meta.env.VITE_API_URL || '';

export const publicApi = axios.create({
  baseURL: `${API_URL}/api/v1/public`,
  headers: {
    'Content-Type': 'application/json',
  },
});

export interface PublicOrganization {
  name: string;
  slug: string;
  timezone: string;
  logo_url?: string;
  address?: string;
  phone?: string;
  email?: string;
  cancellation_policy?: string;
  rescheduling_policy?: string;
  email_required: boolean;
  currency: string;
  time_format: '24h' | '12h';
}

export interface PublicBranch {
  id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  phone?: string;
}

export interface PublicService {
  id: string;
  name: string;
  description?: string;
  duration_minutes: number;
  price: number;
  currency: string;
  category?: string;
  color?: string;
}

export interface PublicEmployee {
  id: string;
  first_name: string;
  last_name: string;
  title?: string;
  avatar_url?: string;
}

export interface PublicTimeSlot {
  start_time: string;
  end_time: string;
  available: boolean;
  employee_ids?: string[];
}

export interface PublicBooking {
  id: string;
  start_time: string;
  end_time: string;
  status: string;
  price: number;
  currency: string;
}

export interface PublicCustomerInfo {
  first_name: string;
  last_name: string;
  phone: string;
  email?: string;
}

export interface CreatePublicBookingRequest {
  branch_id: string;
  service_id: string;
  employee_id?: string;
  start_time: string;
  notes?: string;
  customer: PublicCustomerInfo;
}

export async function getPublicOrganization(slug: string): Promise<PublicOrganization> {
  const { data } = await publicApi.get<PublicOrganization>(`/${slug}`);
  return data;
}

export async function getPublicBranches(slug: string): Promise<PublicBranch[]> {
  const { data } = await publicApi.get<PublicBranch[]>(`/${slug}/branches`);
  return data;
}

export async function getPublicServices(slug: string, branchId: string): Promise<PublicService[]> {
  const { data } = await publicApi.get<PublicService[]>(`/${slug}/services`, {
    params: { branch_id: branchId },
  });
  return data;
}

export async function getPublicEmployees(slug: string, branchId: string, serviceId: string): Promise<PublicEmployee[]> {
  const { data } = await publicApi.get<PublicEmployee[]>(`/${slug}/employees`, {
    params: { branch_id: branchId, service_id: serviceId },
  });
  return data;
}

export async function getPublicAvailability(
  slug: string,
  params: { branch_id: string; service_id: string; date: string; employee_id?: string },
): Promise<PublicTimeSlot[]> {
  const { data } = await publicApi.get<PublicTimeSlot[]>(`/${slug}/availability`, { params });
  return data;
}

export async function createPublicBooking(
  slug: string,
  payload: CreatePublicBookingRequest,
): Promise<PublicBooking> {
  const { data } = await publicApi.post<PublicBooking>(`/${slug}/bookings`, payload);
  return data;
}

export { getApiErrorMessage };
