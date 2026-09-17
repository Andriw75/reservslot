import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { forkJoin } from 'rxjs';
import { environment } from '../../environments/environment';
import type { Ambiente } from '../models/ambiente';
import type { AmbienteInfo } from '../models/ambiente-info';
import type { Categoria } from '../models/categoria';
import type { Imagen } from '../models/imagen';
import type { Ocupacion } from '../models/ocupacion';
import { SseService } from './sse.service';

export interface AmbienteInfoState {
  info: AmbienteInfo | null;
  imagenes: Imagen[];
}

export interface OcupacionesState {
  from: string;
  to: string;
  data: Ocupacion[];
}

const EMPTY_INFO: AmbienteInfoState = { info: null, imagenes: [] };
const EMPTY_OCUPACIONES: OcupacionesState = { from: '', to: '', data: [] };

@Injectable({ providedIn: 'root' })
export class SportslotStoreService {
  private http = inject(HttpClient);
  private sse = inject(SseService);
  private readonly api = environment.API_URL;

  readonly categorias = signal<Categoria[]>([]);
  readonly ambientes = signal<Ambiente[]>([]);
  readonly infos = signal<Record<number, AmbienteInfoState>>({});
  readonly ocupaciones = signal<Record<number, OcupacionesState>>({});

  private coreActive = false;
  private readonly extraTopics = new Map<string, number>();

  constructor() {
    this.sse.events$.subscribe((message) => this.apply(message.event, message.topic, message.data));
  }

  ensureCatalog(): void {
    if (this.coreActive) return;
    this.coreActive = true;
    this.syncTopics();
    this.loadCatalog();
  }

  loadCatalog(): void {
    forkJoin({
      categorias: this.http.get<Categoria[]>(`${this.api}/categorias`),
      ambientes: this.http.get<Ambiente[]>(`${this.api}/ambientes`),
    }).subscribe({
      next: ({ categorias, ambientes }) => {
        this.categorias.set(categorias);
        this.ambientes.set(ambientes);
      },
    });
  }

  loadInfo(ambienteId: number): void {
    this.http.get<AmbienteInfo>(`${this.api}/ambientes/${ambienteId}/info`).subscribe({
      next: (info) => this.upsertInfo(ambienteId, info, this.info(ambienteId).imagenes),
    });
  }

  loadOcupaciones(ambienteId: number, from?: string, to?: string): void {
    let url = `${this.api}/ambientes/${ambienteId}/ocupaciones`;
    if (from && to) {
      url += `?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`;
    }
    this.http.get<OcupacionesState>(url).subscribe({
      next: (range) => this.ocupaciones.update((map) => ({ ...map, [ambienteId]: range })),
    });
  }

  setOcupaciones(ambienteId: number, state: OcupacionesState): void {
    this.ocupaciones.update((map) => ({ ...map, [ambienteId]: state }));
  }

  subscribeTopic(topic: string): () => void {
    this.extraTopics.set(topic, (this.extraTopics.get(topic) ?? 0) + 1);
    this.syncTopics();

    let active = true;
    return () => {
      if (!active) return;
      active = false;
      const count = (this.extraTopics.get(topic) ?? 0) - 1;
      if (count > 0) {
        this.extraTopics.set(topic, count);
      } else {
        this.extraTopics.delete(topic);
      }
      this.syncTopics();
    };
  }

  info(ambienteId: number): AmbienteInfoState {
    return this.infos()[ambienteId] ?? EMPTY_INFO;
  }

  ocupacionesFor(ambienteId: number): OcupacionesState {
    return this.ocupaciones()[ambienteId] ?? EMPTY_OCUPACIONES;
  }

  categoriaName(categoriaId: number): string {
    return (
      this.categorias().find((categoria) => categoria.id === categoriaId)?.nombre ?? 'Sin categoría'
    );
  }

  disconnect(): void {
    this.coreActive = false;
    this.extraTopics.clear();
    this.sse.disconnect();
  }

  private syncTopics(): void {
    const topics = new Set<string>();
    if (this.coreActive) {
      topics.add('categorias');
      topics.add('ambientes');
    }
    for (const [topic, count] of this.extraTopics) {
      if (count > 0) topics.add(topic);
    }
    if (topics.size === 0) {
      this.sse.disconnect();
      return;
    }
    this.sse.connect([...topics]);
  }

  private apply(event: string, topic: string, data: any): void {
    switch (event) {
      case 'snapshot':
        this.applySnapshot(topic, data);
        return;
      case 'categoria.upsert':
        this.upsertCategoria(data as Categoria);
        return;
      case 'categoria.delete':
        this.removeCategoria(Number(data?.id));
        return;
      case 'ambiente.upsert':
        this.upsertAmbiente(data as Ambiente);
        return;
      case 'ambiente.delete':
        this.removeAmbiente(Number(data?.id));
        return;
      case 'ambiente_info.upsert':
        this.applyInfoEvent(topic, data);
        return;
      case 'ocupacion.upsert':
        this.upsertOcupacion(topic, data as Ocupacion);
        return;
      case 'ocupacion.delete':
        this.removeOcupacion(topic, Number(data?.id));
        return;
    }
  }

  private applySnapshot(topic: string, data: any): void {
    if (topic === 'categorias') {
      this.categorias.set((data as Categoria[]) ?? []);
      return;
    }
    if (topic === 'ambientes') {
      this.ambientes.set((data as Ambiente[]) ?? []);
      return;
    }

    const ambienteId = this.ambienteIdFromTopic(topic);
    if (ambienteId == null) return;

    if (topic.endsWith(':info')) {
      const payload = data as { info?: AmbienteInfo; imagenes?: Imagen[] } | null;
      if (payload?.info) this.upsertInfo(ambienteId, payload.info, payload.imagenes ?? []);
      return;
    }
    if (topic.endsWith(':ocupaciones')) {
      const payload = data as OcupacionesState | null;
      if (payload) this.ocupaciones.update((map) => ({ ...map, [ambienteId]: payload }));
    }
  }

  private applyInfoEvent(topic: string, data: any): void {
    const ambienteId = this.ambienteIdFromTopic(topic);
    if (ambienteId == null) return;
    const payload = data as { info?: AmbienteInfo; imagenes?: Imagen[] } | null;
    if (!payload?.info) return;
    this.upsertInfo(ambienteId, payload.info, payload.imagenes ?? []);
  }

  private upsertInfo(ambienteId: number, info: AmbienteInfo, imagenes: Imagen[]): void {
    this.infos.update((map) => ({ ...map, [ambienteId]: { info, imagenes } }));
  }

  private upsertOcupacion(topic: string, ocupacion: Ocupacion): void {
    const ambienteId = this.ambienteIdFromTopic(topic);
    if (ambienteId == null || !ocupacion) return;
    this.ocupaciones.update((map) => {
      const current = map[ambienteId] ?? EMPTY_OCUPACIONES;
      const data = current.data
        .filter((item) => item.id !== ocupacion.id)
        .concat(ocupacion)
        .sort((a, b) => a.inicio.localeCompare(b.inicio));
      return { ...map, [ambienteId]: { ...current, data } };
    });
  }

  private removeOcupacion(topic: string, ocupacionId: number): void {
    if (Number.isNaN(ocupacionId)) return;
    const ambienteId = this.ambienteIdFromTopic(topic);
    if (ambienteId == null) return;
    this.ocupaciones.update((map) => {
      const current = map[ambienteId];
      if (!current) return map;
      return {
        ...map,
        [ambienteId]: { ...current, data: current.data.filter((item) => item.id !== ocupacionId) },
      };
    });
  }

  private upsertCategoria(categoria: Categoria): void {
    if (!categoria) return;
    this.categorias.update((list) => {
      const exists = list.some((item) => item.id === categoria.id);
      return exists
        ? list.map((item) => (item.id === categoria.id ? categoria : item))
        : [...list, categoria];
    });
  }

  private removeCategoria(id: number): void {
    if (Number.isNaN(id)) return;
    this.categorias.update((list) => list.filter((item) => item.id !== id));
  }

  private upsertAmbiente(ambiente: Ambiente): void {
    if (!ambiente) return;
    this.ambientes.update((list) => {
      const exists = list.some((item) => item.id === ambiente.id);
      return exists
        ? list.map((item) => (item.id === ambiente.id ? ambiente : item))
        : [...list, ambiente];
    });
  }

  private removeAmbiente(id: number): void {
    if (Number.isNaN(id)) return;
    this.ambientes.update((list) => list.filter((item) => item.id !== id));
  }

  private ambienteIdFromTopic(topic: string): number | null {
    const parts = topic.split(':');
    if (parts.length < 3 || parts[0] !== 'ambiente') return null;
    const id = Number(parts[1]);
    return Number.isNaN(id) ? null : id;
  }
}
