import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';
import { MENU, type MenuItem } from './menu.config';
import { SidebarMenuItemComponent } from './sidebar-menu-item';
import { UserSolidIcon } from '../../../common/icons/user-solid.icon';
import { LogoutIcon } from '../../../common/icons/logout.icon';
import { SpinnerIcon } from '../../../common/icons/spinner.icon';

@Component({
  selector: 'app-sidebar',
  templateUrl: './sidebar.html',
  styleUrl: './sidebar.css',
  imports: [SidebarMenuItemComponent, UserSolidIcon, LogoutIcon, SpinnerIcon],
})
export class SidebarComponent {
  private auth = inject(AuthService);
  private router = inject(Router);
  private store = inject(SportslotStoreService);

  user = this.auth.user;
  menu = MENU;

  isCollapsed = signal(false);
  isLoggingOut = signal(false);

  toggleSidebar(): void {
    this.isCollapsed.update((value) => !value);
  }

  logout(): void {
    this.store.disconnect();
    this.isLoggingOut.set(true);
    this.auth.logout().subscribe({
      next: () => {
        this.auth.user.set(null);
      },
      error: () => {
        this.auth.user.set(null);
        this.isLoggingOut.set(false);
        this.router.navigate(['/login']);
      },
      complete: () => {
        this.isLoggingOut.set(false);
        this.router.navigate(['/login']);
      },
    });
  }

  filterMenu(items: MenuItem[]): MenuItem[] {
    return items
      .map((item) => (item.submenu ? { ...item, submenu: this.filterMenu(item.submenu) } : item))
      .filter(
        (item) =>
          !item.requiredPermission ||
          (this.user()?.permisos?.includes(item.requiredPermission) ?? false) ||
          (item.submenu && item.submenu.length > 0),
      );
  }
}
