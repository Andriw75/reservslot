import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';
import { AuthService } from '../services/auth.service';

export const loginGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.user()) {
    router.navigate(['/dashboard']);
    return of(false);
  }

  return auth.me().pipe(
    map((user) => {
      auth.user.set(user);
      router.navigate(['/dashboard']);
      return false;
    }),
    catchError(() => of(true)),
  );
};
