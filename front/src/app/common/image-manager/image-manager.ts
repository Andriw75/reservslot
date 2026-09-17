import { Component, effect, inject, input, OnDestroy, output, signal } from '@angular/core';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';
import { firstValueFrom } from 'rxjs';
import type { Observable } from 'rxjs';
import { imagenUrl } from '../../models/imagen';
import type { Imagen } from '../../models/imagen';
import { ToastService } from '../../services/toast.service';

export type CardStatus = 'uploaded' | 'pending-upload' | 'pending-delete';

export interface ImagenCard {
  key: number;
  id: number | null;
  url: string;
  file?: File;
  status: CardStatus;
}

export interface ImagenesApi {
  uploadImagenes(id: number, files: File[]): Observable<Imagen[]>;
  deleteImagen(id: number, imagenId: number): Observable<void>;
  reorderImagenes(id: number, ids: number[]): Observable<Imagen[]>;
}

const STATUS_LABEL: Record<CardStatus, string> = {
  uploaded: 'En servidor',
  'pending-upload': 'Por subir',
  'pending-delete': 'Por eliminar',
};

@Component({
  selector: 'app-image-manager',
  templateUrl: './image-manager.html',
  styleUrl: './image-manager.css',
  imports: [DragDropModule],
})
export class ImageManagerComponent implements OnDestroy {
  private toast = inject(ToastService);

  entityId = input.required<number>();
  imagenes = input<Imagen[]>([]);
  api = input.required<ImagenesApi>();

  imagesApplied = output<void>();

  cards = signal<ImagenCard[]>([]);
  applying = signal(false);
  zoneActive = signal(false);
  orderDirty = signal(false);

  private nextKey = 1;
  private prevCards: ImagenCard[] = [];
  private lastImagenes: Imagen[] | null = null;

  constructor() {
    effect(() => {
      const imagenes = this.imagenes();
      if (imagenes === this.lastImagenes) return;
      this.lastImagenes = imagenes;
      this.syncFromServer(imagenes);
    });
  }

  ngOnDestroy(): void {
    this.revokeLocalUrls(this.cards());
  }

  // Refresca las tarjetas "en servidor" sin perder las pendientes de subir o
  // eliminar (por ejemplo cuando llega un evento SSE con la lista actualizada).
  private syncFromServer(imagenes: Imagen[]): void {
    const serverIds = new Set(imagenes.map((img) => img.id));
    const kept: ImagenCard[] = [];

    for (const card of this.cards()) {
      if (card.status === 'pending-upload') {
        kept.push(card);
        continue;
      }
      if (card.status === 'pending-delete') {
        if (card.id !== null && serverIds.has(card.id)) {
          kept.push(card);
        } else if (card.url.startsWith('blob:')) {
          URL.revokeObjectURL(card.url);
        }
        continue;
      }
      const current = imagenes.find((img) => img.id === card.id);
      if (current) {
        kept.push({ ...card, url: imagenUrl(current) });
      } else if (card.url.startsWith('blob:')) {
        URL.revokeObjectURL(card.url);
      }
    }

    const knownIds = new Set(kept.map((card) => card.id));
    const added: ImagenCard[] = [...imagenes]
      .sort((a, b) => a.orden - b.orden)
      .filter((img) => !knownIds.has(img.id))
      .map((img) => ({
        key: this.nextKey++,
        id: img.id,
        url: imagenUrl(img),
        status: 'uploaded' as CardStatus,
      }));

    const next = [...kept, ...added];
    this.prevCards = next;
    this.cards.set(next);
  }

  private toCards(imagenes: Imagen[]): ImagenCard[] {
    return [...imagenes]
      .sort((a, b) => a.orden - b.orden)
      .map((img) => ({
        key: this.nextKey++,
        id: img.id,
        url: imagenUrl(img),
        status: 'uploaded' as CardStatus,
      }));
  }

  private revokeLocalUrls(cards: ImagenCard[]): void {
    for (const card of cards) {
      if (card.file && card.url.startsWith('blob:')) {
        URL.revokeObjectURL(card.url);
      }
    }
  }

  onFilesSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    const files = input.files ? Array.from(input.files) : [];
    if (!files.length) return;
    this.addPending(files);
    input.value = '';
  }

  dropCard(event: CdkDragDrop<ImagenCard[]>): void {
    if (event.previousIndex === event.currentIndex) return;
    const list = [...this.cards()];
    moveItemInArray(list, event.previousIndex, event.currentIndex);
    this.cards.set(list);
    this.prevCards = list;
    this.orderDirty.set(true);
  }

  removeOrRestore(card: ImagenCard): void {
    this.cards.update((list) => {
      if (card.status === 'pending-upload') {
        if (card.url.startsWith('blob:')) URL.revokeObjectURL(card.url);
        return list.filter((item) => item.key !== card.key);
      }
      if (card.status === 'uploaded') {
        this.orderDirty.set(true);
        return list.map((item) =>
          item.key === card.key ? { ...item, status: 'pending-delete' as CardStatus } : item,
        );
      }
      this.orderDirty.set(true);
      return list.map((item) =>
        item.key === card.key ? { ...item, status: 'uploaded' as CardStatus } : item,
      );
    });
  }

  statusLabel(card: ImagenCard): string {
    return STATUS_LABEL[card.status];
  }

  onZoneDragOver(event: DragEvent): void {
    if (this.applying() || !this.hasFiles(event)) return;
    event.preventDefault();
    event.dataTransfer!.dropEffect = 'copy';
    this.zoneActive.set(true);
  }

  onZoneDragLeave(): void {
    this.zoneActive.set(false);
  }

  onZoneDrop(event: DragEvent): void {
    if (this.applying() || !this.hasFiles(event)) return;
    event.preventDefault();
    this.zoneActive.set(false);
    const files = event.dataTransfer ? Array.from(event.dataTransfer.files) : [];
    if (!files.length) return;
    this.addPending(files);
  }

  private addPending(files: File[]): void {
    const newCards: ImagenCard[] = files.map((file) => ({
      key: this.nextKey++,
      id: null,
      url: URL.createObjectURL(file),
      file,
      status: 'pending-upload',
    }));
    this.cards.update((list) => [...list, ...newCards]);
    this.orderDirty.set(true);
  }

  private hasFiles(event: DragEvent): boolean {
    const types = event.dataTransfer?.types ?? [];
    return Array.from(types).includes('Files');
  }

  get pendingUploads(): number {
    return this.cards().filter((card) => card.status === 'pending-upload').length;
  }

  get pendingDeletes(): number {
    return this.cards().filter((card) => card.status === 'pending-delete').length;
  }

  get hasPending(): boolean {
    return this.pendingUploads + this.pendingDeletes > 0 || this.orderDirty();
  }

  get pendingSummary(): string {
    const parts: string[] = [];
    if (this.pendingUploads) parts.push(`${this.pendingUploads} subir`);
    if (this.pendingDeletes) parts.push(`${this.pendingDeletes} eliminar`);
    if (this.orderDirty() && !parts.length) parts.push('reordenar');
    return parts.length ? `Aplicar cambios (${parts.join(' · ')})` : 'Aplicar cambios';
  }

  async applyImages(): Promise<void> {
    const id = this.entityId();
    this.applying.set(true);
    try {
      const current = this.cards();
      const toUpload = current.filter((card) => card.status === 'pending-upload');
      const toDelete = current.filter((card) => card.status === 'pending-delete');
      const uploadedKeys = new Set(toUpload.map((card) => card.key));
      const deletedKeys = new Set(toDelete.map((card) => card.key));
      const shouldReorder = this.orderDirty();

      for (const card of toDelete) {
        await firstValueFrom(this.api().deleteImagen(id, card.id!));
      }
      if (toUpload.length) {
        await firstValueFrom(
          this.api().uploadImagenes(
            id,
            toUpload.map((card) => card.file!),
          ),
        );
      }

      if (shouldReorder) {
        const ids = this.cards()
          .filter((card) => card.status === 'uploaded' && card.id !== null)
          .map((card) => card.id!);
        if (ids.length) {
          await firstValueFrom(this.api().reorderImagenes(id, ids));
        }
      }

      this.revokeLocalUrls(toUpload);
      const next = this.cards().filter(
        (card) => !uploadedKeys.has(card.key) && !deletedKeys.has(card.key),
      );
      this.prevCards = next;
      this.cards.set(next);
      this.orderDirty.set(false);
      this.toast.success('Cambios de imágenes aplicados');
      this.imagesApplied.emit();
    } catch {
      this.toast.error('Error al aplicar cambios de imágenes');
    } finally {
      this.applying.set(false);
    }
  }

  discardChanges(): void {
    if (this.applying()) return;
    this.revokeLocalUrls(this.cards());
    const cards = this.toCards(this.imagenes());
    this.prevCards = cards;
    this.cards.set(cards);
    this.orderDirty.set(false);
    this.toast.success('Cambios descartados');
  }
}
