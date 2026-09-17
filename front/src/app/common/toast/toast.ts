import { Component, inject } from '@angular/core';
import { ToastService } from '../../services/toast.service';

@Component({
  selector: 'app-toast',
  templateUrl: './toast.html',
  styleUrl: './toast.css',
})
export class ToastComponent {
  protected toast = inject(ToastService);
}
