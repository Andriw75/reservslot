import { Component, computed, effect, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { OcupacionesService } from '../../../services/ocupaciones.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { ConfirmService } from '../../../services/confirm.service';
import { OcupacionModalComponent } from './modal';
import type { Ocupacion } from '../../../models/ocupacion';
import { LIMA_OFFSET, formatLocalInput, formatOcupacionRange, weekRange } from './ocupacion-utils';

@Component({
  selector: 'app-ocupaciones-list',
  templateUrl: './list.html',
  styleUrl: './list.css',
  imports: [FormsModule, OcupacionModalComponent],
})
export class OcupacionesListComponent {
  private service = inject(OcupacionesService);
  private store = inject(SportslotStoreService);
  private toast = inject(ToastService);
  private confirm = inject(ConfirmService);

  ambientes = this.store.ambientes;
  selectedAmbienteId = signal<number | null>(null);
  range = signal(weekRange(new Date()));
  loading = signal(false);

  showModal = signal(false);
  selected = signal<Ocupacion | null>(null);
  deleting = signal<number | null>(null);

  ocupaciones = computed(() => {
    const id = this.selectedAmbienteId();
    return id == null ? [] : this.store.ocupacionesFor(id).data;
  });

  selectedAmbiente = computed(() =>
    this.ambientes().find((ambiente) => ambiente.id === this.selectedAmbienteId()),
  );

  constructor() {
    effect((onCleanup) => {
      const id = this.selectedAmbienteId();
      if (id == null) return;
      const fromIso = `${this.range().from}T00:00:00${LIMA_OFFSET}`;
      const toIso = `${this.range().to}T00:00:00${LIMA_OFFSET}`;
      this.service.list(id, fromIso, toIso).subscribe({
        next: (result) => this.store.setOcupaciones(id, result),
        error: () => this.toast.error('Error al cargar ocupaciones'),
      });
      const unsubscribe = this.store.subscribeTopic(`ambiente:${id}:ocupaciones`);
      onCleanup(unsubscribe);
    });

    effect(() => {
      const ambientes = this.ambientes();
      if (this.selectedAmbienteId() == null && ambientes.length) {
        this.selectedAmbienteId.set(ambientes[0].id);
      }
    });
  }

  onAmbienteChange(value: number): void {
    this.selectedAmbienteId.set(Number(value));
  }

  onRangeChange(): void {
    this.range.set({ ...this.range() });
  }

  openCreate(): void {
    this.selected.set(null);
    this.showModal.set(true);
  }

  openEdit(ocupacion: Ocupacion): void {
    this.selected.set(ocupacion);
    this.showModal.set(true);
  }

  onSaved(): void {
    const id = this.selectedAmbienteId();
    if (id != null) this.store.loadOcupaciones(id, this.fromIso(), this.toIso());
  }

  async confirmDelete(ocupacion: Ocupacion): Promise<void> {
    const ok = await this.confirm.confirm(
      'Eliminar ocupación',
      `¿Seguro de eliminar la ocupación ${this.formatRange(ocupacion)}?`,
      'danger',
    );
    if (!ok) return;

    this.deleting.set(ocupacion.id);
    this.service.delete(ocupacion.id).subscribe({
      next: () => {
        this.toast.success('Ocupación eliminada');
        this.deleting.set(null);
        this.onSaved();
      },
      error: () => {
        this.deleting.set(null);
        this.toast.error('Error al eliminar la ocupación');
      },
    });
  }

  formatRange(ocupacion: Ocupacion): string {
    return formatOcupacionRange(ocupacion);
  }

  toLocalInput(value: string): string {
    return formatLocalInput(value);
  }

  private fromIso(): string {
    return `${this.range().from}T00:00:00${LIMA_OFFSET}`;
  }

  private toIso(): string {
    return `${this.range().to}T00:00:00${LIMA_OFFSET}`;
  }
}
