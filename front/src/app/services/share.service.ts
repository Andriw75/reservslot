import { DOCUMENT } from '@angular/common';
import { Injectable, inject } from '@angular/core';

export interface ShareData {
  title?: string;
  text?: string;
  url: string;
}

@Injectable({ providedIn: 'root' })
export class ShareService {
  private document = inject(DOCUMENT);

  // Enlace público del catálogo de ambientes. Con `id` el portal queda con ese
  // ambiente resaltado; sin `id` abre el catálogo tal cual.
  // Se apoya en <base href> para respetar la subruta en la que se publica la SPA
  // (p. ej. /reservslot/) en lugar de asumir siempre la raíz del dominio.
  ambienteUrl(id?: number | null): string {
    const base = this.document.baseURI;
    return id ? new URL(`?ambiente=${id}`, base).toString() : base;
  }

  canNativeShare(): boolean {
    return typeof navigator !== 'undefined' && typeof navigator.share === 'function';
  }

  // En móviles/táctiles conviene la hoja nativa del sistema; en escritorio el
  // usuario espera copiar el enlace directamente.
  prefersNativeShare(): boolean {
    if (!this.canNativeShare()) return false;
    try {
      return window.matchMedia('(pointer: coarse)').matches;
    } catch {
      return false;
    }
  }

  async nativeShare(data: ShareData): Promise<boolean> {
    if (!this.canNativeShare()) return false;
    try {
      await navigator.share(data);
      return true;
    } catch {
      return false;
    }
  }

  async copy(text: string): Promise<boolean> {
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(text);
        return true;
      } catch {
        // Continúa con el fallback.
      }
    }
    return this.copyFallback(text);
  }

  private copyFallback(text: string): boolean {
    try {
      const area = document.createElement('textarea');
      area.value = text;
      area.setAttribute('readonly', '');
      area.style.position = 'fixed';
      area.style.top = '-1000px';
      area.style.opacity = '0';
      document.body.appendChild(area);
      area.select();
      area.setSelectionRange(0, text.length);
      const ok = document.execCommand('copy');
      document.body.removeChild(area);
      return ok;
    } catch {
      return false;
    }
  }
}
