import { Component, inject, signal } from '@angular/core';
import { AmbientesService } from '../../../services/ambientes.service';
import { ConfirmService } from '../../../services/confirm.service';
import { ShareService } from '../../../services/share.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { ShareButtonComponent } from '../../../common/share/share';
import { AmbienteModalComponent } from './modal';
import { AmbienteInfoModalComponent } from './info-modal';
import type { Ambiente } from '../../../models/ambiente';

@Component({
  selector: 'app-ambientes-list',
  templateUrl: './list.html',
  styleUrl: './list.css',
  imports: [AmbienteModalComponent, AmbienteInfoModalComponent, ShareButtonComponent],
})
export class AmbientesListComponent {
  private store = inject(SportslotStoreService);
  private service = inject(AmbientesService);
  private toast = inject(ToastService);
  private confirm = inject(ConfirmService);
  private share = inject(ShareService);

  ambientes = this.store.ambientes;
  categorias = this.store.categorias;

  showModal = signal(false);
  showInfoModal = signal(false);
  selected = signal<Ambiente | null>(null);
  deleting = signal<number | null>(null);

  categoriaName(id: number): string {
    return this.categorias().find((categoria) => categoria.id === id)?.nombre ?? 'Sin categoría';
  }

  ambienteUrl(id: number): string {
    return this.share.ambienteUrl(id);
  }

  openCreate(): void {
    this.selected.set(null);
    this.showModal.set(true);
  }

  openEdit(ambiente: Ambiente): void {
    this.selected.set(ambiente);
    this.showModal.set(true);
  }

  openInfo(ambiente: Ambiente): void {
    this.selected.set(ambiente);
    this.store.loadInfo(ambiente.id);
    this.showInfoModal.set(true);
  }

  onSaved(): void {
    this.store.loadCatalog();
  }

  async confirmDelete(ambiente: Ambiente): Promise<void> {
    const ok = await this.confirm.confirm(
      'Eliminar ambiente',
      `¿Seguro de eliminar el ambiente "${ambiente.nombre}"? Se eliminarán su información e imágenes.`,
      'danger',
    );
    if (!ok) return;

    this.deleting.set(ambiente.id);
    this.service.delete(ambiente.id).subscribe({
      next: () => this.onDeleted(),
      error: async (err) => {
        if (err?.status === 409) {
          const total = Number(err?.error?.ocupaciones ?? 0);
          this.deleting.set(null);
          const confirmed = await this.confirm.confirm(
            'Ambiente con ocupaciones',
            `"${ambiente.nombre}" tiene ${total} ocupación(es) registrada(s). ` +
              `Si continúas se eliminarán también esas ocupaciones. ¿Deseas continuar?`,
            'danger',
          );
          if (!confirmed) return;
          this.deleting.set(ambiente.id);
          this.service.delete(ambiente.id, true).subscribe({
            next: () => this.onDeleted(),
            error: () => {
              this.deleting.set(null);
              this.toast.error('Error al eliminar ambiente');
            },
          });
          return;
        }
        this.deleting.set(null);
        this.toast.error('Error al eliminar ambiente');
      },
    });
  }

  private onDeleted(): void {
    this.toast.success('Ambiente eliminado');
    this.deleting.set(null);
    this.store.loadCatalog();
  }
}
