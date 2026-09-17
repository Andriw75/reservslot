import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { SpinnerIcon } from '../../common/icons/spinner.icon';

@Component({
  selector: 'app-login',
  templateUrl: './login.html',
  styleUrl: './login.css',
  imports: [RouterLink, SpinnerIcon],
})
export class LoginComponent {
  private auth = inject(AuthService);
  private router = inject(Router);

  username = signal('');
  password = signal('');
  error = signal('');
  isLoading = signal(false);

  handleSubmit(e: Event): void {
    e.preventDefault();

    const user = this.username().trim();
    const pass = this.password().trim();

    if (!user || !pass) {
      const missing: string[] = [];
      if (!user) missing.push('Usuario');
      if (!pass) missing.push('Contraseña');
      this.error.set(`Complete los campos: ${missing.join(', ')}`);
      return;
    }

    this.isLoading.set(true);
    this.error.set('');

    this.auth.login(user, pass).subscribe({
      next: (u) => {
        this.auth.user.set(u);
        this.router.navigate(['/dashboard']);
      },
      error: () => {
        this.isLoading.set(false);
        this.error.set('Usuario o contraseña incorrectos');
      },
    });
  }
}
