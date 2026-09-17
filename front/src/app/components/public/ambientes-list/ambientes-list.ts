import { Component, computed, effect, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import type { Ambiente } from '../../../models/ambiente';
import type { Categoria } from '../../../models/categoria';

export interface CategoriaGroup {
  categoria: Categoria | null;
  ambientes: Ambiente[];
}

@Component({
  selector: 'app-public-ambientes-list',
  templateUrl: './ambientes-list.html',
  styleUrl: './ambientes-list.css',
  imports: [RouterLink],
})
export class AmbientesListComponent {
  private store = inject(SportslotStoreService);

  categorias = this.store.categorias;
  ambientes = this.store.ambientes;

  groups = computed<CategoriaGroup[]>(() => {
    const ambientes = this.ambientes().filter((ambiente) => ambiente.activo);
    const groups: CategoriaGroup[] = [];

    for (const categoria of this.categorias().filter((item) => item.activo)) {
      groups.push({
        categoria,
        ambientes: ambientes.filter((ambiente) => ambiente.categoria_id === categoria.id),
      });
    }

    const known = new Set(this.categorias().map((categoria) => categoria.id));
    const orphan = ambientes.filter((ambiente) => !known.has(ambiente.categoria_id));
    if (orphan.length) {
      groups.push({ categoria: null, ambientes: orphan });
    }
    return groups;
  });

  constructor() {
    this.store.ensureCatalog();
    effect((onCleanup) => {
      const unsubscribe = this.store.subscribeTopic('categorias');
      onCleanup(unsubscribe);
    });
  }

  categoryColor(group: CategoriaGroup): string {
    return group.categoria?.color || 'var(--primary)';
  }

  hasAmbientes(group: CategoriaGroup): boolean {
    return group.ambientes.length > 0;
  }
}
