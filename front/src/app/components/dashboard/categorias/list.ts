import { Component, inject, signal } from '@angular/core';
import { CategoriasService } from '../../../services/categorias.service';
import { ConfirmService } from '../../../services/confirm.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ToastService } from '../../../services/toast.service';
import { CategoriaModalComponent } from './modal';
import type { Categoria } from '../../../models/categoria';

@Component({
  selector: 'app-categorias-list',
  templateUrl: './list.html',
  imports: [CategoriaModalComponent],
})
export class CategoriasListComponent {
  private store = inject(SportslotStoreService);
  private service = inject(CategoriasService);
  private toast = inject(ToastService);
  private confirm = inject(ConfirmService);

  categorias = this.store.categorias;
  showModal = signal(false);
  selected = signal<Categoria | null>(null);
  deleting = signal<number | null>(null);

  openCreate(): void {
    this.selected.set(null);
    this.showModal.set(true);
  }

  openEdit(categoria: Categoria): void {
    this.selected.set(categoria);
    this.showModal.set(true);
  }

  onSaved(): void {
    this.store.loadCatalog();
  }

  async confirmDelete(categoria: Categoria): Promise<void> {
    const ok = await this.confirm.confirm(
      'Eliminar categoría',
      `¿Seguro de eliminar la categoría "${categoria.nombre}"?`,
      'danger',
    );
    if (!ok) return;

    this.deleting.set(categoria.id);
    this.service.delete(categoria.id).subscribe({
      next: () => {
        this.toast.success('Categoría eliminada');
        this.deleting.set(null);
        this.store.loadCatalog();
      },
      error: (error) => {
        this.deleting.set(null);
        this.toast.error(
          error?.status === 409
            ? 'No se puede eliminar: la categoría tiene ambientes asociados'
            : 'Error al eliminar categoría',
        );
      },
    });
  }
}
