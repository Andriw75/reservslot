import { Component, Input } from '@angular/core';

@Component({
  selector: 'icon-user-solid',
  template: `<svg
    xmlns="http://www.w3.org/2000/svg"
    [attr.width]="size"
    [attr.height]="size"
    viewBox="0 0 24 24"
    fill="currentColor"
    stroke="none"
  >
    <path
      d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"
    />
  </svg>`,
  styles: `
    :host {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  `,
})
export class UserSolidIcon {
  @Input() size = '20';
}
