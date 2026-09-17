import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AmbientesService } from '../../../services/ambientes.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { ImageManagerComponent } from '../../../common/image-manager/image-manager';
import type { ImagenesApi } from '../../../common/image-manager/image-manager';
import type { Ambiente } from '../../../models/ambiente';

const DEFAULT_OPTIONS: Record<string, boolean> = {
  techado: false,
  iluminacion: false,
  vestuarios: false,
  gradas: false,
};

@Component({
  selector: 'app-ambiente-info-modal',
  templateUrl: './info-modal.html',
  styleUrl: './info-modal.css',
  imports: [FormsModule, ImageManagerComponent],
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
  };

  saving = signal(false);
  error = signal('');
  dataLoaded = signal(false);

  form: {
    ancho: number | null;
    alto: number | null;
    aforo: number | null;
    descripcion: string;
    opciones_adicionales: Record<string, boolean>;
  } = {
    ancho: null,
    alto: null,
    aforo: null,
    descripcion: '',
    opciones_adicionales: { ...DEFAULT_OPTIONS },
  };

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
        opciones_adicionales: {
          ...DEFAULT_OPTIONS,
          ...(data['opciones_adicionales'] ?? {}),
        },
      };
      this.dataLoaded.set(true);
    });
  }

  optionKeys(): string[] {
    return Object.keys(this.form.opciones_adicionales);
  }

  close(): void {
    this.onClose.emit();
  }

  onSubmit(): void {
    this.saving.set(true);
    this.error.set('');

    const data: Record<string, unknown> = {
      ancho: this.form.ancho,
      alto: this.form.alto,
      aforo: this.form.aforo,
      descripcion: this.form.descripcion,
      opciones_adicionales: this.form.opciones_adicionales,
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
