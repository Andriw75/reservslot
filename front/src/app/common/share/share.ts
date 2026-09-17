import { Component, inject, input, signal } from '@angular/core';
import { ShareService } from '../../services/share.service';
import { ToastService } from '../../services/toast.service';

export type ShareVariant = 'button' | 'icon' | 'flat' | 'ghost';

@Component({
  selector: 'app-share-button',
  templateUrl: './share.html',
  styleUrl: './share.css',
})
export class ShareButtonComponent {
  private share = inject(ShareService);
  private toast = inject(ToastService);

  url = input.required<string>();
  title = input('ReservSlot');
  text = input('Mira este ambiente');
  label = input('Compartir');
  variant = input<ShareVariant>('button');
  disabled = input(false);

  copied = signal(false);
  manualUrl = signal<string | null>(null);

  async onShare(): Promise<void> {
    if (this.disabled()) return;

    const url = this.url();

    if (this.share.prefersNativeShare()) {
      const shared = await this.share.nativeShare({
        title: this.title(),
        text: this.text(),
        url,
      });
      if (shared) return;
    }

    await this.copyLink();
  }

  async copyLink(): Promise<void> {
    const url = this.url();
    const ok = await this.share.copy(url);

    if (ok) {
      this.manualUrl.set(null);
      this.copied.set(true);
      this.toast.success('Enlace copiado');
      setTimeout(() => this.copied.set(false), 2000);
      return;
    }

    this.manualUrl.set(url);
  }

  closeManual(): void {
    this.manualUrl.set(null);
  }

  selectAll(event: Event): void {
    (event.target as HTMLInputElement).select();
  }
}
