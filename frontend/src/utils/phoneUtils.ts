import { z } from 'zod';

export const PHONE_PLACEHOLDER = '+370 123 12345';
export const PHONE_FORMAT_MESSAGE = `Use international format, e.g. ${PHONE_PLACEHOLDER}`;

const E164_REGEX = /^\+[1-9]\d{6,14}$/;
const LT_DISPLAY_REGEX = /^\+370 \d{3} \d{5}$/;

export function normalizePhone(raw: string): string {
  let phone = raw.trim().replace(/[\s\-()]/g, '');

  if (phone.startsWith('00')) {
    phone = `+${phone.slice(2)}`;
  } else if (!phone.startsWith('+') && phone.startsWith('0')) {
    phone = `+370${phone.slice(1)}`;
  } else if (!phone.startsWith('+')) {
    phone = `+${phone}`;
  }

  return phone;
}

export function formatPhoneDisplay(raw: string): string {
  if (!raw) return '';
  const normalized = normalizePhone(raw);
  const ltMatch = normalized.match(/^\+370(\d{3})(\d{5})$/);
  if (ltMatch) {
    return `+370 ${ltMatch[1]} ${ltMatch[2]}`;
  }
  return normalized;
}

export function isValidPhoneInput(raw: string): boolean {
  const trimmed = raw.trim();
  if (!trimmed) return false;
  if (LT_DISPLAY_REGEX.test(trimmed)) return true;
  return E164_REGEX.test(normalizePhone(trimmed));
}

export function isValidE164(phone: string): boolean {
  return E164_REGEX.test(phone);
}

export const optionalPhoneField = z
  .string()
  .optional()
  .or(z.literal(''))
  .refine((value) => !value || isValidPhoneInput(value), { message: PHONE_FORMAT_MESSAGE })
  .transform((value) => (value ? normalizePhone(value) : ''));

export const requiredPhoneField = z
  .string()
  .min(1, 'Phone is required')
  .refine(isValidPhoneInput, { message: PHONE_FORMAT_MESSAGE })
  .transform((value) => normalizePhone(value));
