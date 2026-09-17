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

export function weekRange(date = new Date()): { from: string; to: string; days: Date[] } {
  const weekday = (date.getDay() + 6) % 7;
  const monday = shiftDate(date, -weekday);
  const days = Array.from({ length: 7 }, (_, index) => shiftDate(monday, index));
  return { from: toDateKey(monday), to: toDateKey(shiftDate(monday, 7)), days };
}

export function formatRangeLabel(from: string, to: string): string {
  if (!from || !to) return '';
  const start = new Date(`${from}T00:00:00`);
  const end = shiftDate(new Date(`${to}T00:00:00`), -1);
  const format = (date: Date) =>
    `${String(date.getDate()).padStart(2, '0')}/${String(date.getMonth() + 1).padStart(2, '0')}/${date.getFullYear()}`;
  return `${format(start)} al ${format(end)}`;
}
