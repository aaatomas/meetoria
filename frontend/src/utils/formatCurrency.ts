const currencySymbolCache = new Map<string, string>();

export function getCurrencySymbol(currency: string): string {
  const code = currency.toUpperCase();
  const cached = currencySymbolCache.get(code);
  if (cached) {
    return cached;
  }

  const symbol = new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: code,
    currencyDisplay: 'narrowSymbol',
  })
    .formatToParts(0)
    .find((part) => part.type === 'currency')?.value ?? code;

  currencySymbolCache.set(code, symbol);
  return symbol;
}

export function formatPrice(amount: number, currency: string): string {
  const code = currency?.trim().toUpperCase();
  const safeCode = code && code.length === 3 ? code : 'EUR';
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: safeCode,
      currencyDisplay: 'narrowSymbol',
    }).format(amount);
  } catch {
    return `${amount} ${safeCode}`;
  }
}
