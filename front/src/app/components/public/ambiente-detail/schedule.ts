import { toDateKey } from './time';

export const DAY_NAMES = ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo'];
export const DAY_SHORT = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom'];

export const DEFAULT_START_HOUR = 7;
export const DEFAULT_END_HOUR = 22;

// Alto en píxeles de cada fila de hora. Debe coincidir con `.schedule-cell`
// y `.schedule-hour` en ambiente-detail.css.
export const HOUR_HEIGHT = 34;

// Un bloque con menos de una hora de duración no tiene espacio para mostrar
// hora + rango, así que solo muestra la hora; si dura menos de media hora no
// muestra texto y la información queda en el tooltip del hover.
const COMPACT_HOURS = 1;
const TINY_HOURS = 0.5;

export interface ScheduleSegment {
  key: string;
  ocupacionId: number;
  dayIndex: number;
  top: number;
  height: number;
  label: string;
  startLabel: string;
  endLabel: string;
  isStart: boolean;
  isEnd: boolean;
  compact: boolean;
  tiny: boolean;
  tooltip: string[];
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
  startHour: number;
  endHour: number;
  segments: ScheduleSegment[];
  crossesDays: Set<string>;
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

function startOfDay(date: Date): Date {
  const copy = new Date(date.getTime());
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function timeLabel(date: Date): string {
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  return `${hours}:${minutes}`;
}

function decimalHour(date: Date): number {
  return date.getHours() + date.getMinutes() / 60;
}

function dateLabel(date: Date): string {
  const day = DAY_SHORT[(date.getDay() + 6) % 7];
  const dd = String(date.getDate()).padStart(2, '0');
  const mm = String(date.getMonth() + 1).padStart(2, '0');
  return `${day} ${dd}/${mm}`;
}

function durationLabel(minutes: number): string {
  const hours = Math.floor(minutes / 60);
  const mins = Math.round(minutes % 60);
  if (hours && mins) return `${hours} h ${mins} min`;
  if (hours) return `${hours} h`;
  return `${mins} min`;
}

// Reparte cada ocupación en segmentos por día, recortando los extremos al
// rango visible del grid. Un rango que cruza medianoche genera un segmento por
// día, todos con el mismo `key` para poder resaltarlos juntos en el hover.
export function buildWeek(ocupaciones: OcupacionLike[], reference = new Date()): WeekSchedule {
  const weekday = (reference.getDay() + 6) % 7;
  const monday = startOfDay(shiftDate(reference, -weekday));
  const today = toDateKey(reference);

  const days: WeekDay[] = DAY_NAMES.map((name, index) => {
    const date = shiftDate(monday, index);
    const key = toDateKey(date);
    return { date: key, name, short: DAY_SHORT[index], isToday: key === today };
  });

  const dayBounds = days.map((day) => {
    const from = new Date(`${day.date}T00:00:00`).getTime();
    return { from, to: from + 24 * 3600 * 1000 };
  });

  const ranges: { ocupacionId: number; inicio: Date; fin: Date }[] = [];
  let startHour = DEFAULT_START_HOUR;
  let endHour = DEFAULT_END_HOUR;

  for (const ocupacion of ocupaciones) {
    const inicio = new Date(ocupacion.inicio);
    const fin = new Date(ocupacion.fin);
    if (Number.isNaN(inicio.getTime()) || Number.isNaN(fin.getTime())) continue;
    if (fin.getTime() <= inicio.getTime()) continue;

    ranges.push({ ocupacionId: ocupacion.id, inicio, fin });

    const sameDay = toDateKey(inicio) === toDateKey(new Date(fin.getTime() - 1));
    if (sameDay) {
      startHour = Math.min(startHour, Math.floor(decimalHour(inicio)));
      endHour = Math.max(endHour, Math.ceil(decimalHour(fin)));
    } else {
      startHour = 0;
      endHour = 24;
    }
  }

  startHour = Math.max(0, Math.min(startHour, 23));
  endHour = Math.max(startHour + 1, Math.min(24, endHour));

  const hours: number[] = [];
  for (let hour = startHour; hour < endHour; hour++) {
    hours.push(hour);
  }

  const span = endHour - startHour;
  const segments: ScheduleSegment[] = [];
  const crossesDays = new Set<string>();

  for (const range of ranges) {
    const ocupacionId = range.ocupacionId;
    const key = `o${ocupacionId}`;
    const inicioLabel = timeLabel(range.inicio);
    const finLabel = timeLabel(range.fin);
    const totalMinutes = (range.fin.getTime() - range.inicio.getTime()) / 60000;

    for (let dayIndex = 0; dayIndex < days.length; dayIndex++) {
      const bounds = dayBounds[dayIndex];
      const from = Math.max(range.inicio.getTime(), bounds.from);
      const to = Math.min(range.fin.getTime(), bounds.to);
      if (to <= from) continue;

      const startDecimal = (from - bounds.from) / 3600000;
      const endDecimal = (to - bounds.from) / 3600000;

      const top = ((startDecimal - startHour) / span) * 100;
      const height = ((endDecimal - startDecimal) / span) * 100;
      if (height <= 0) continue;

      const isStart = from === range.inicio.getTime();
      const isEnd = to === range.fin.getTime();
      if (!isStart || !isEnd) crossesDays.add(key);

      const durationHours = (to - from) / 3600000;

      const dayLabel = DAY_NAMES[dayIndex];
      const rangeLabel =
        isStart && isEnd
          ? `${dayLabel} ${inicioLabel} - ${finLabel}`
          : `${dayLabel} ${timeLabel(new Date(from))} - ${timeLabel(new Date(to))}`;

      segments.push({
        key,
        ocupacionId,
        dayIndex,
        top,
        height,
        label: timeLabel(new Date(from)),
        startLabel: isStart ? inicioLabel : timeLabel(new Date(from)),
        endLabel: isEnd ? finLabel : timeLabel(new Date(to)),
        isStart,
        isEnd,
        compact: durationHours < COMPACT_HOURS,
        tiny: durationHours < TINY_HOURS,
        tooltip: [
          rangeLabel,
          `Ocupación: ${dateLabel(range.inicio)} ${inicioLabel} - ${finLabel}`,
          `Duración: ${durationLabel(totalMinutes)}`,
        ],
      });
    }
  }

  for (const segment of segments) {
    if (crossesDays.has(segment.key)) segment.tooltip.push('Continúa en otro día');
  }

  return { days, hours, startHour, endHour, segments, crossesDays };
}
