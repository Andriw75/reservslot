import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

const SKIP_PATHS = ['/auth/token', '/auth/me'];

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const skip = SKIP_PATHS.some((p) => req.url.includes(p));
  if (skip) {
    return next(req);
  }

  const auth = inject(AuthService);
  const router = inject(Router);

  return next(req).pipe(
    catchError((err) => {
      if (err instanceof HttpErrorResponse && err.status === 401 && auth.user()) {
        auth.user.set(null);
        router.navigate(['/login']);
      }
      return throwError(() => err);
    }),
  );
};
