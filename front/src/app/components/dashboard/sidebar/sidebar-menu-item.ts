import { Component, Input, signal, inject } from '@angular/core';
import { NgComponentOutlet } from '@angular/common';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import type { MenuItem } from './menu.config';

@Component({
  selector: 'sidebar-menu-item',
  templateUrl: './sidebar-menu-item.html',
  styleUrl: './sidebar-menu-item.css',
  imports: [NgComponentOutlet, RouterLink, RouterLinkActive, SidebarMenuItemComponent],
})
export class SidebarMenuItemComponent {
  private router = inject(Router);

  @Input({ required: true }) item!: MenuItem;
  @Input() collapsed = false;

  isOpen = signal(false);

  toggle(): void {
    this.isOpen.update((v) => !v);
  }

  isActive(path?: string): boolean {
    return this.router.url === path;
  }
}
