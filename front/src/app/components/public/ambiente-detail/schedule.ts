import { toDateKey } from './time';

export const DAY_NAMES = ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo'];
export const DAY_SHORT = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom'];
export const START_HOUR = 7;
export const END_HOUR = 22;

export interface ScheduleBlock {
  id: number;
  dayIndex: number;
  left: number;
  width: number;
  label: string;
  subtitle: string;
}

export interface WeekDay {
  date: string;
  name: string;
  short: string;
  isToday: boolean;
}

export interface WeekSchedule {
  days: WeekDay[];
  hours: number[];
  blocks: ScheduleBlock[];
}

interface OcupacionLike {
  id: number;
  inicio: string;
  fin: string;
}

function shiftDate(date: Date, days: number): Date {
  const copy = new Date(date.getTime());
  copy.setDate(copy.getDate() + days);
  return copy;
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function timeLabel(date: Date): string {
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${hours}:${minutes}`;
}

export function buildWeek(ocupaciones: OcupacionLike[], reference = new Date()): WeekSchedule {
  const weekday = (reference.getDay() + 6) % 7;
  const monday = shiftDate(reference, -weekday);
  const today = toDateKey(reference);

  const days: WeekDay[] = DAY_NAMES.map((name, index) => {
    const date = shiftDate(monday, index);
    const key = toDateKey(date);
    return { date: key, name, short: DAY_SHORT[index], isToday: key === today };
  });

  const hours: number[] = [];
  for (let hour = START_HOUR; hour <= END_HOUR; hour++) {
    hours.push(hour);
  }

  const span = END_HOUR - START_HOUR;
  const blocks: ScheduleBlock[] = [];

  for (const ocupacion of ocupaciones) {
    const inicio = new Date(ocupacion.inicio);
    const fin = new Date(ocupacion.fin);
    if (Number.isNaN(inicio.getTime()) || Number.isNaN(fin.getTime())) continue;

    const dayIndex = days.findIndex((day) => day.date === toDateKey(inicio));
    if (dayIndex < 0) continue;

    const start = inicio.getHours() + inicio.getMinutes() / 60;
    const end = fin.getHours() + fin.getMinutes() / 60;
    const from = clamp(start - START_HOUR, 0, span);
    const to = clamp(end - START_HOUR, 0, span);
    const width = Math.max(to - from, 0.4);

    blocks.push({
      id: ocupacion.id,
      dayIndex,
      left: (from / span) * 100,
      width: (width / span) * 100,
      label: timeLabel(inicio),
      subtitle: `${days[dayIndex].short} ${timeLabel(inicio)}-${timeLabel(fin)}`,
    });
  }

  return { days, hours, blocks };
}
