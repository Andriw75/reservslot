import { Injectable, signal } from '@angular/core';

export type ConfirmType = 'danger' | 'warning' | 'primary';

export interface ConfirmState {
  visible: boolean;
  title: string;
  message: string;
  type: ConfirmType;
  resolve: ((value: boolean) => void) | null;
}

@Injectable({ providedIn: 'root' })
export class ConfirmService {
  #state = signal<ConfirmState>({
    visible: false,
    title: '',
    message: '',
    type: 'primary',
    resolve: null,
  });
  readonly state = this.#state.asReadonly();

  confirm(title: string, message: string, type: ConfirmType = 'primary'): Promise<boolean> {
    return new Promise((resolve) => {
      this.#state.set({ visible: true, title, message, type, resolve });
    });
  }

  confirmAction(): void {
    const s = this.#state();
    if (s.resolve) {
      s.resolve(true);
    }
    this.#state.set({ visible: false, title: '', message: '', type: 'primary', resolve: null });
  }

  cancel(): void {
    const s = this.#state();
    if (s.resolve) {
      s.resolve(false);
    }
    this.#state.set({ visible: false, title: '', message: '', type: 'primary', resolve: null });
  }
}
