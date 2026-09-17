import { Injectable, signal } from '@angular/core';

export interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info';
}

@Injectable({ providedIn: 'root' })
export class ToastService {
  toasts = signal<Toast[]>([]);
  private nextId = 0;

  success(message: string): void {
    this.add(message, 'success');
  }

  error(message: string): void {
    this.add(message, 'error');
  }

  info(message: string): void {
    this.add(message, 'info');
  }

  remove(id: number): void {
    this.toasts.update((t) => t.filter((toast) => toast.id !== id));
  }

  private add(message: string, type: Toast['type']): void {
    const id = ++this.nextId;
    this.toasts.update((t) => [...t, { id, message, type }]);
    setTimeout(() => this.remove(id), 4000);
  }
}
