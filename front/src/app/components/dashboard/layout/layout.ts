import { Component, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from '../sidebar/sidebar';
import { SportslotStoreService } from '../../../services/sportslot-store.service';

@Component({
  selector: 'app-dashboard-layout',
  templateUrl: './layout.html',
  styleUrl: './layout.css',
  imports: [RouterOutlet, SidebarComponent],
})
export class DashboardLayoutComponent {
  constructor() {
    inject(SportslotStoreService).ensureCatalog();
  }
}
