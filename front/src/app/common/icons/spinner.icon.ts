import { Component, Input } from '@angular/core';

@Component({
  selector: 'icon-spinner',
  template: `<svg
    xmlns="http://www.w3.org/2000/svg"
    [attr.width]="size"
    [attr.height]="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path d="M21 12a9 9 0 1 1-6.219-8.56" />
  </svg>`,
  styles: `
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      animation: spin 0.7s linear infinite;
    }
    @keyframes spin {
      to {
        transform: rotate(360deg);
      }
    }
  `,
})
export class SpinnerIcon {
  @Input() size = '20';
}
