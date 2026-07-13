const E164_REGEX = /^\+[1-9]\d{6,14}$/;

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

export function isValidE164(phone: string): boolean {
  return E164_REGEX.test(phone);
}
