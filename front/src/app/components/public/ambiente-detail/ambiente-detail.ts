import { Component, computed, effect, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { imagenUrl } from '../../../models/imagen';
import type { AmbienteInfoData } from '../../../models/ambiente-info';
import type { Imagen } from '../../../models/imagen';
import { buildWeek } from './schedule';
import type { ScheduleBlock } from './schedule';
import { formatRangeLabel } from './time';

interface InfoRow {
  label: string;
  value: string;
}

@Component({
  selector: 'app-public-ambiente-detail',
  templateUrl: './ambiente-detail.html',
  styleUrl: './ambiente-detail.css',
  imports: [RouterLink],
})
export class AmbienteDetailComponent {
  private store = inject(SportslotStoreService);

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

  schedule = computed(() => buildWeek(this.ocupaciones().data));
  weekLabel = computed(() => formatRangeLabel(this.ocupaciones().from, this.ocupaciones().to));

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

  options = computed<InfoRow[]>(() => {
    const data: AmbienteInfoData = this.info().info?.data ?? {};
    const options = data.opciones_adicionales ?? {};
    return Object.entries(options).map(([label, value]) => ({
      label,
      value: value ? 'Sí' : 'No',
    }));
  });

  currentImageUrl = computed(() => {
    const image = this.currentImage();
    return image ? imagenUrl(image) : '';
  });

  private readonly blocksByDay = computed(() => {
    const grouped = new Map<number, ScheduleBlock[]>();
    for (const block of this.schedule().blocks) {
      const list = grouped.get(block.dayIndex) ?? [];
      list.push(block);
      grouped.set(block.dayIndex, list);
    }
    return grouped;
  });

  imageUrl(image: Imagen): string {
    return imagenUrl(image);
  }

  blocksForDay(dayIndex: number): ScheduleBlock[] {
    return this.blocksByDay().get(dayIndex) ?? [];
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
      if (this.store.ambientes().length && !this.ambiente()) {
        this.notFound.set(true);
      }
    });
  }

  selectImage(index: number): void {
    this.activeImage.set(index);
  }
}
