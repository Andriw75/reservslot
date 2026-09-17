import { Component, Input } from '@angular/core';

@Component({
  selector: 'icon-calendar',
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
    <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
    <line x1="16" y1="2" x2="16" y2="6" />
    <line x1="8" y1="2" x2="8" y2="6" />
    <line x1="3" y1="10" x2="21" y2="10" />
  </svg>`,
  styles: `
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  `,
})
export class CalendarIcon {
  @Input() size = '20';
}
