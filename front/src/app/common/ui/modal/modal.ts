import { Component, HostListener, inject, input, output } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { A11yModule } from '@angular/cdk/a11y';

let nextModalId = 0;

@Component({
  selector: 'app-modal',
  templateUrl: './modal.html',
  styleUrl: './modal.css',
  imports: [A11yModule],
})
export class ModalComponent {
  private document = inject(DOCUMENT);
  private previousOverflow = '';

  title = input('');
  subtitle = input('');
  width = input(520);
  showClose = input(true);
  closeOnBackdrop = input(true);
  closed = output<void>();

  readonly titleId = `app-modal-title-${nextModalId++}`;

  constructor() {
    const body = this.document.body;
    this.previousOverflow = body.style.overflow;
    body.style.overflow = 'hidden';
  }

  ngOnDestroy(): void {
    this.document.body.style.overflow = this.previousOverflow;
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    this.close();
  }

  onBackdrop(): void {
    if (this.closeOnBackdrop()) {
      this.close();
    }
  }

  close(): void {
    this.closed.emit();
  }
}
