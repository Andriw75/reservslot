import { Component, computed, effect, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ShareService } from '../../../services/share.service';
import { ShareButtonComponent } from '../../../common/share/share';
import { imagenArchivoUrl } from '../../../models/imagen';
import type { Ambiente } from '../../../models/ambiente';
import type { Categoria } from '../../../models/categoria';

export interface CategoriaGroup {
  categoria: Categoria | null;
  ambientes: Ambiente[];
}

interface Paso {
  titulo: string;
  texto: string;
}

@Component({
  selector: 'app-public-ambientes-list',
  templateUrl: './ambientes-list.html',
  styleUrl: './ambientes-list.css',
  imports: [RouterLink, ShareButtonComponent],
})
export class AmbientesListComponent {
  private store = inject(SportslotStoreService);
  private share = inject(ShareService);

  // Permite enlaces directos al catálogo con un ambiente resaltado:
  //   /?ambiente=3
  ambiente = input<string>();

  categorias = this.store.categorias;
  ambientes = this.store.ambientes;

  highlightedId = signal<number | null>(null);

  private alreadyHighlighted: number | null = null;

  readonly pasos: Paso[] = [
    {
      titulo: 'Elige la categoría',
      texto: 'Canchas, piscinas, salas y más. Cada categoría agrupa sus ambientes.',
    },
    {
      titulo: 'Revisa la disponibilidad',
      texto: 'Mira el horario semanal de cada ambiente, actualizado en tiempo real.',
    },
    {
      titulo: 'Confirma tu horario',
      texto: 'Selecciona el día y la hora libre que más te convenga, sin llamadas.',
    },
  ];

  totalAmbientes = computed(() => this.ambientes().filter((a) => a.activo).length);
  totalCategorias = computed(() => this.groups().filter((g) => g.ambientes.length > 0).length);

  groups = computed<CategoriaGroup[]>(() => {
    const activos = this.ambientes().filter((ambiente) => ambiente.activo);
    const activas = this.categorias().filter((categoria) => categoria.activo);

    const grupos: CategoriaGroup[] = activas
      .map((categoria) => ({
        categoria,
        ambientes: activos.filter((ambiente) => ambiente.categoria_id === categoria.id),
      }))
      .filter((grupo) => grupo.ambientes.length > 0);

    const conocidas = new Set(activas.map((categoria) => categoria.id));
    const huerfanos = activos.filter((ambiente) => !conocidas.has(ambiente.categoria_id));
    if (huerfanos.length) {
      grupos.push({ categoria: null, ambientes: huerfanos });
    }
    return grupos;
  });

  constructor() {
    this.store.ensureCatalog();
    effect((onCleanup) => {
      const unsubscribe = this.store.subscribeTopic('categorias');
      onCleanup(unsubscribe);
    });

    // Enlace compartido: cuando el catálogo esté listo, baja hasta la tarjeta
    // del ambiente pedido y la resalta unos segundos.
    effect((onCleanup) => {
      const raw = this.ambiente();
      const id = raw ? Number(raw) : Number.NaN;
      if (Number.isNaN(id)) return;
      if (!this.store.catalogReady()) return;
      if (!this.ambientes().some((item) => item.id === id)) return;
      if (this.alreadyHighlighted === id) return;

      this.alreadyHighlighted = id;
      this.highlightedId.set(id);

      const scrollTimer = setTimeout(() => this.scrollToAmbiente(id), 80);
      const clearTimer = setTimeout(() => this.highlightedId.set(null), 4000);

      onCleanup(() => {
        clearTimeout(scrollTimer);
        clearTimeout(clearTimer);
      });
    });
  }

  ambienteUrl(id: number): string {
    return this.share.ambienteUrl(id);
  }

  categoryColor(group: CategoriaGroup): string {
    return group.categoria?.color || 'var(--primary)';
  }

  ambienteImage(ambiente: Ambiente): string | null {
    return ambiente.imagen_principal ? imagenArchivoUrl(ambiente.imagen_principal) : null;
  }

  scrollTo(event: Event, id: string): void {
    event.preventDefault();
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  private scrollToAmbiente(id: number): void {
    document
      .getElementById(`ambiente-${id}`)
      ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}
