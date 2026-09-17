import { Component, computed, effect, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import { SportslotStoreService } from '../../../services/sportslot-store.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.html',
  styleUrl: './home.css',
  imports: [RouterLink],
})
export class HomeComponent {
  private store = inject(SportslotStoreService);
  auth = inject(AuthService);

  categorias = this.store.categorias;
  ambientes = this.store.ambientes;
  ocupacionesSemana = computed(() =>
    this.store
      .ambientes()
      .reduce((total, ambiente) => total + this.store.ocupacionesFor(ambiente.id).data.length, 0),
  );

  constructor() {
    this.store.ensureCatalog();
    effect((onCleanup) => {
      const unsubscribes = this.store.ambientes().map((ambiente) => {
        this.store.loadOcupaciones(ambiente.id);
        return this.store.subscribeTopic(`ambiente:${ambiente.id}:ocupaciones`);
      });
      onCleanup(() => unsubscribes.forEach((unsubscribe) => unsubscribe()));
    });
  }
}
