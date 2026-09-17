import { Injectable } from '@angular/core';

export interface ShareData {
  title?: string;
  text?: string;
  url: string;
}

@Injectable({ providedIn: 'root' })
export class ShareService {
  // Enlace público del catálogo de ambientes. Con `id` el portal queda con ese
  // ambiente resaltado; sin `id` abre el catálogo tal cual.
  ambienteUrl(id?: number | null): string {
    const base = `${window.location.origin}/`;
    return id ? `${base}?ambiente=${id}` : base;
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
