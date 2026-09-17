import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';
import { loginGuard } from './guards/login.guard';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    loadComponent: () =>
      import('./components/public/ambientes-list/ambientes-list').then(
        (m) => m.AmbientesListComponent,
      ),
  },
  {
    path: 'ambientes/:id',
    loadComponent: () =>
      import('./components/public/ambiente-detail/ambiente-detail').then(
        (m) => m.AmbienteDetailComponent,
      ),
  },
  {
    path: 'login',
    canActivate: [loginGuard],
    loadComponent: () => import('./components/login/login').then((m) => m.LoginComponent),
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/dashboard/layout/layout').then((m) => m.DashboardLayoutComponent),
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./components/dashboard/home/home').then((m) => m.HomeComponent),
      },
      {
        path: 'categorias',
        loadComponent: () =>
          import('./components/dashboard/categorias/list').then((m) => m.CategoriasListComponent),
      },
      {
        path: 'ambientes',
        loadComponent: () =>
          import('./components/dashboard/ambientes/list').then((m) => m.AmbientesListComponent),
      },
      {
        path: 'ocupaciones',
        loadComponent: () =>
          import('./components/dashboard/ocupaciones/list').then((m) => m.OcupacionesListComponent),
      },
    ],
  },
  {
    path: '**',
    redirectTo: '',
  },
];
