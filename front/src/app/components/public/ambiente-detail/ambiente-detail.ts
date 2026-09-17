import { Component, computed, effect, inject, input, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { imagenUrl } from '../../../models/imagen';
import type { AmbienteInfoData } from '../../../models/ambiente-info';
import type { Imagen } from '../../../models/imagen';
import { buildWeek } from './schedule';
import type { ScheduleSegment } from './schedule';
import { formatRangeLabel } from './time';

interface InfoRow {
  label: string;
  value: string;
}

interface AttributeRow {
  clave: string;
  valor: string;
}

interface TooltipState {
  x: number;
  y: number;
  lines: string[];
}

@Component({
  selector: 'app-public-ambiente-detail',
  templateUrl: './ambiente-detail.html',
  styleUrl: './ambiente-detail.css',
  imports: [RouterLink],
})
export class AmbienteDetailComponent {
  private store = inject(SportslotStoreService);
  private toast = inject(ToastService);
  private router = inject(Router);

  id = input.required<string>();

  ambienteId = computed(() => Number(this.id()));
  ambiente = computed(() => this.store.ambientes().find((item) => item.id === this.ambienteId()));
  categoria = computed(() => {
    const current = this.ambiente();
    return current
      ? (this.store.categorias().find((item) => item.id === current.categoria_id) ?? null)
      : null;
  });
  info = computed(() => this.store.info(this.ambienteId()));
  ocupaciones = computed(() => this.store.ocupacionesFor(this.ambienteId()));

  activeImage = signal(0);
  notFound = signal(false);
  removed = signal(false);

  private hadAmbiente = false;

  schedule = computed(() => buildWeek(this.ocupaciones().data));
  weekLabel = computed(() => formatRangeLabel(this.ocupaciones().from, this.ocupaciones().to));

  hoveredKey = signal<string | null>(null);
  tooltip = signal<TooltipState | null>(null);

  images = computed<Imagen[]>(() => [...this.info().imagenes].sort((a, b) => a.orden - b.orden));

  currentImage = computed<Imagen | null>(() => {
    const images = this.images();
    if (!images.length) return null;
    return images[Math.min(this.activeImage(), images.length - 1)];
  });

  infoRows = computed<InfoRow[]>(() => {
    const data = this.info().info?.data ?? {};
    const rows: InfoRow[] = [];
    const add = (label: string, value: unknown, unit = '') => {
      if (value === null || value === undefined || value === '') return;
      rows.push({ label, value: `${value}${unit}` });
    };
    add('Ancho', data.ancho, ' m');
    add('Alto', data.alto, ' m');
    add('Aforo', data.aforo);
    return rows;
  });

  description = computed(() => {
    const value = (this.info().info?.data ?? {})['descripcion'];
    return typeof value === 'string' ? value : '';
  });

  attributes = computed<AttributeRow[]>(() => readAttributes(this.info().info?.data ?? {}));

  currentImageUrl = computed(() => {
    const image = this.currentImage();
    return image ? imagenUrl(image) : '';
  });

  private readonly segmentsByDay = computed(() => {
    const grouped = new Map<number, ScheduleSegment[]>();
    for (const segment of this.schedule().segments) {
      const list = grouped.get(segment.dayIndex) ?? [];
      list.push(segment);
      grouped.set(segment.dayIndex, list);
    }
    return grouped;
  });

  imageUrl(image: Imagen): string {
    return imagenUrl(image);
  }

  segmentsForDay(dayIndex: number): ScheduleSegment[] {
    return this.segmentsByDay().get(dayIndex) ?? [];
  }

  constructor() {
    effect((onCleanup) => {
      const id = this.ambienteId();
      if (Number.isNaN(id)) return;
      this.notFound.set(false);
      this.activeImage.set(0);
      this.store.loadCatalog();
      this.store.loadInfo(id);
      this.store.loadOcupaciones(id);
      const unsubscribers = [
        this.store.subscribeTopic(`ambiente:${id}:info`),
        this.store.subscribeTopic(`ambiente:${id}:ocupaciones`),
        this.store.subscribeTopic('categorias'),
        this.store.subscribeTopic('ambientes'),
      ];
      onCleanup(() => unsubscribers.forEach((unsubscribe) => unsubscribe()));
    });

    effect(() => {
      const current = this.ambiente();

      if (current) {
        this.hadAmbiente = true;
        this.removed.set(false);
        return;
      }
      if (!this.store.catalogReady()) return;

      if (this.hadAmbiente) {
        // El ambiente existía y dejó de estar disponible: fue eliminado en vivo.
        this.removed.set(true);
        this.toast.error('Este ambiente fue eliminado por otro usuario');
        void this.router.navigate(['/']);
        return;
      }
      this.notFound.set(true);
    });
  }

  selectImage(index: number): void {
    this.activeImage.set(index);
  }

  isHighlighted(segment: ScheduleSegment): boolean {
    return this.hoveredKey() === segment.key;
  }

  onSegmentEnter(event: MouseEvent, segment: ScheduleSegment): void {
    this.hoveredKey.set(segment.key);
    this.updateTooltip(event, segment);
  }

  onSegmentMove(event: MouseEvent, segment: ScheduleSegment): void {
    if (this.hoveredKey() !== segment.key) return;
    this.updateTooltip(event, segment);
  }

  onSegmentLeave(): void {
    this.hoveredKey.set(null);
    this.tooltip.set(null);
  }

  private updateTooltip(event: MouseEvent, segment: ScheduleSegment): void {
    this.tooltip.set({ x: event.clientX, y: event.clientY, lines: segment.tooltip });
  }
}

function readAttributes(data: AmbienteInfoData): AttributeRow[] {
  const attrs = data['atributos'];
  if (Array.isArray(attrs)) {
    return attrs
      .map((item) => item as { clave?: unknown; valor?: unknown })
      .filter((item) => item && typeof item.clave === 'string')
      .map((item) => ({ clave: item.clave as string, valor: String(item.valor ?? '') }));
  }

  const legacy = data.opciones_adicionales;
  if (legacy && typeof legacy === 'object') {
    return Object.entries(legacy).map(([clave, valor]) => ({
      clave,
      valor: valor ? 'Sí' : 'No',
    }));
  }

  return [];
}
