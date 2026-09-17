import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';
import { AmbientesService } from '../../../services/ambientes.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { ImageManagerComponent } from '../../../common/image-manager/image-manager';
import type { ImagenesApi } from '../../../common/image-manager/image-manager';
import type { Ambiente } from '../../../models/ambiente';
import type { AtributoItem } from '../../../models/ambiente-info';

@Component({
  selector: 'app-ambiente-info-modal',
  templateUrl: './info-modal.html',
  styleUrl: './info-modal.css',
  imports: [FormsModule, DragDropModule, ImageManagerComponent],
})
export class AmbienteInfoModalComponent {
  private service = inject(AmbientesService);
  private store = inject(SportslotStoreService);
  private toast = inject(ToastService);

  ambiente = input.required<Ambiente>();
  onClose = output<void>();

  infoApi: ImagenesApi = {
    uploadImagenes: (id, files) => this.service.uploadImagenes(id, files),
    deleteImagen: (id, imagenId) => this.service.deleteImagen(id, imagenId),
    reorderImagenes: (id, ids) => this.service.reorderImagenes(id, ids),
  };

  saving = signal(false);
  error = signal('');
  dataLoaded = signal(false);

  form: {
    ancho: number | null;
    alto: number | null;
    aforo: number | null;
    descripcion: string;
  } = {
    ancho: null,
    alto: null,
    aforo: null,
    descripcion: '',
  };

  atributos = signal<AtributoItem[]>([]);

  info = computed(() => this.store.info(this.ambiente().id));

  constructor() {
    effect((onCleanup) => {
      const ambiente = this.ambiente();
      this.dataLoaded.set(false);
      this.store.loadInfo(ambiente.id);
      const unsubscribe = this.store.subscribeTopic(`ambiente:${ambiente.id}:info`);
      onCleanup(unsubscribe);
    });

    effect(() => {
      const info = this.info().info;
      if (!info) return;
      const data = info.data ?? {};
      const toNumber = (value: unknown): number | null => {
        if (value === null || value === undefined || value === '') return null;
        const parsed = Number(value);
        return Number.isNaN(parsed) ? null : parsed;
      };
      this.form = {
        ancho: toNumber(data['ancho']),
        alto: toNumber(data['alto']),
        aforo: toNumber(data['aforo']),
        descripcion: typeof data['descripcion'] === 'string' ? data['descripcion'] : '',
      };
      this.atributos.set(readAtributos(data));
      this.dataLoaded.set(true);
    });
  }

  addAtributo(): void {
    this.atributos.update((list) => [...list, { clave: '', valor: '' }]);
  }

  removeAtributo(index: number): void {
    this.atributos.update((list) => list.filter((_, i) => i !== index));
  }

  updateAtributo(index: number, field: 'clave' | 'valor', value: string): void {
    this.atributos.update((list) =>
      list.map((item, i) => (i === index ? { ...item, [field]: value } : item)),
    );
  }

  dropAtributo(event: CdkDragDrop<AtributoItem[]>): void {
    if (event.previousIndex === event.currentIndex) return;
    const list = [...this.atributos()];
    moveItemInArray(list, event.previousIndex, event.currentIndex);
    this.atributos.set(list);
  }

  close(): void {
    this.onClose.emit();
  }

  onImagesApplied(): void {
    this.store.loadInfo(this.ambiente().id);
  }

  onSubmit(): void {
    this.saving.set(true);
    this.error.set('');

    const atributos = this.atributos()
      .map((item) => ({ clave: item.clave.trim(), valor: item.valor.trim() }))
      .filter((item) => item.clave.length > 0 || item.valor.length > 0);

    const data: Record<string, unknown> = {
      ancho: this.form.ancho,
      alto: this.form.alto,
      aforo: this.form.aforo,
      descripcion: this.form.descripcion,
      atributos,
    };

    this.service.updateInfo(this.ambiente().id, { data, activo: true }).subscribe({
      next: () => {
        this.saving.set(false);
        this.toast.success('Información del ambiente guardada');
        this.store.loadInfo(this.ambiente().id);
      },
      error: () => {
        this.saving.set(false);
        this.error.set('Error al guardar la información del ambiente');
      },
    });
  }
}

function readAtributos(data: Record<string, unknown>): AtributoItem[] {
  const attrs = data['atributos'];
  if (Array.isArray(attrs)) {
    return attrs
      .map((item) => item as { clave?: unknown; valor?: unknown })
      .filter((item) => item && typeof item.clave === 'string')
      .map((item) => ({ clave: String(item.clave), valor: String(item.valor ?? '') }));
  }

  const legacy = data['opciones_adicionales'];
  if (legacy && typeof legacy === 'object') {
    return Object.entries(legacy as Record<string, unknown>).map(([clave, valor]) => ({
      clave,
      valor: valor ? 'Sí' : 'No',
    }));
  }

  return [];
}
