import { Component, HostListener, computed, effect, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { ShareService } from '../../../services/share.service';
import { ShareButtonComponent } from '../../../common/share/share';
import { RevealDirective } from '../../../common/ui/reveal/reveal.directive';
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

interface Beneficio {
  key: 'vivo' | 'sin-llamadas' | 'sin-costo' | 'categorias';
  titulo: string;
  texto: string;
}

interface Faq {
  pregunta: string;
  respuesta: string;
}

@Component({
  selector: 'app-public-ambientes-list',
  templateUrl: './ambientes-list.html',
  styleUrl: './ambientes-list.css',
  imports: [RouterLink, ShareButtonComponent, RevealDirective],
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
  scrolled = signal(false);
  openFaq = signal<number | null>(0);

  private alreadyHighlighted: number | null = null;

  readonly pasos: Paso[] = [
    {
      titulo: 'Elige tu ambiente',
      texto: 'Canchas, piscinas, salas y más, agrupados por categoría.',
    },
    {
      titulo: 'Consulta el horario en vivo',
      texto: 'Mira el calendario semanal real de cada ambiente.',
    },
    {
      titulo: 'Coordina tu reserva',
      texto: 'Confirma el horario disponible con la administración del recinto.',
    },
  ];

  readonly beneficios: Beneficio[] = [
    {
      key: 'vivo',
      titulo: 'Disponibilidad en vivo',
      texto: 'El calendario se actualiza al instante: nada de horarios desactualizados.',
    },
    {
      key: 'sin-llamadas',
      titulo: 'Sin llamadas ni idas y vueltas',
      texto: 'Revisa todos los horarios desde el móvil, cuando quieras.',
    },
    {
      key: 'sin-costo',
      titulo: 'Consulta sin costo',
      texto: 'Ver la disponibilidad no cuesta nada ni requiere registrarte.',
    },
    {
      key: 'categorias',
      titulo: 'Todo en un solo lugar',
      texto: 'Canchas, piscinas, salas y más, ordenados por categoría.',
    },
  ];

  readonly faqs: Faq[] = [
    {
      pregunta: '¿Necesito crear una cuenta?',
      respuesta:
        'No. El portal es público: puedes ver categorías, ambientes y horarios sin registrarte.',
    },
    {
      pregunta: '¿La reserva se confirma en la web?',
      respuesta:
        'Aquí consultas la disponibilidad en vivo. La reserva se coordina y confirma con la administración del recinto.',
    },
    {
      pregunta: '¿Tiene algún costo consultar?',
      respuesta: 'No, consultar la disponibilidad es completamente gratis.',
    },
    {
      pregunta: '¿La disponibilidad está actualizada?',
      respuesta:
        'Sí. El horario se actualiza en tiempo real, así que lo que ves es lo que está libre.',
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

  @HostListener('window:scroll')
  onWindowScroll(): void {
    this.scrolled.set(window.scrollY > 8);
  }

  ambienteUrl(id: number): string {
    return this.share.ambienteUrl(id);
  }

  categoryColor(group: CategoriaGroup): string {
    return group.categoria?.color || 'var(--primary)';
  }

  categoryAnchor(group: CategoriaGroup): string {
    return `categoria-${group.categoria?.id ?? 'otros'}`;
  }

  ambienteImage(ambiente: Ambiente): string | null {
    return ambiente.imagen_principal ? imagenArchivoUrl(ambiente.imagen_principal) : null;
  }

  toggleFaq(index: number): void {
    this.openFaq.update((current) => (current === index ? null : index));
  }

  scrollTo(event: Event, id: string): void {
    event.preventDefault();
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  scrollToCategory(event: Event, group: CategoriaGroup): void {
    event.preventDefault();
    document
      .getElementById(this.categoryAnchor(group))
      ?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  private scrollToAmbiente(id: number): void {
    document
      .getElementById(`ambiente-${id}`)
      ?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}
