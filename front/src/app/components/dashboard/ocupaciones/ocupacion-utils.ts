export const LIMA_OFFSET = '-05:00';

function shiftDate(date: Date, days: number): Date {
  const copy = new Date(date.getTime());
  copy.setDate(copy.getDate() + days);
  return copy;
}

export function toDateKey(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function weekRange(date: Date): { from: string; to: string } {
  const weekday = (date.getDay() + 6) % 7;
  const monday = shiftDate(date, -weekday);
  const nextMonday = shiftDate(monday, 7);
  return { from: toDateKey(monday), to: toDateKey(nextMonday) };
}

export function formatLocalInput(value: string | undefined): string {
  return value ? value.slice(0, 16) : '';
}

export function toLimaIso(localDateTime: string): string {
  if (!localDateTime) return '';
  return localDateTime.length === 16 ? `${localDateTime}:00${LIMA_OFFSET}` : localDateTime;
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const day = String(date.getDate()).padStart(2, '0');
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${day}/${month}/${date.getFullYear()} ${hours}:${minutes}`;
}

export function formatOcupacionRange(ocupacion: { inicio: string; fin: string }): string {
  return `${formatDateTime(ocupacion.inicio)} - ${formatDateTime(ocupacion.fin)}`;
}
