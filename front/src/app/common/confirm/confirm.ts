import { Component, inject } from '@angular/core';
import { ConfirmService } from '../../services/confirm.service';
import { ModalComponent } from '../ui/modal/modal';

@Component({
  selector: 'app-confirm',
  templateUrl: './confirm.html',
  styleUrl: './confirm.css',
  imports: [ModalComponent],
})
export class ConfirmComponent {
  protected confirm = inject(ConfirmService);
}
