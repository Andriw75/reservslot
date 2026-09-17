import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';
import { AuthService } from '../services/auth.service';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.user()) {
    return of(true);
  }

  return auth.me().pipe(
    map((user) => {
      auth.user.set(user);
      return true;
    }),
    catchError(() => {
      router.navigate(['/login']);
      return of(false);
    }),
  );
};
