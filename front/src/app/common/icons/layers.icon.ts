import { Component, Input } from '@angular/core';

@Component({
  selector: 'icon-layers',
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
    <polygon points="12 2 2 7 12 12 22 7 12 2" />
    <polyline points="2 17 12 22 22 17" />
    <polyline points="2 12 12 17 22 12" />
  </svg>`,
  styles: `
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  `,
})
export class LayersIcon {
  @Input() size = '20';
}
