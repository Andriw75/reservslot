import { Component, inject } from '@angular/core';
import { ConfirmService } from '../../services/confirm.service';

@Component({
  selector: 'app-confirm',
  templateUrl: './confirm.html',
  styleUrl: './confirm.css',
})
export class ConfirmComponent {
  protected confirm = inject(ConfirmService);
}
