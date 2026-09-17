import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { OcupacionesService } from '../../../services/ocupaciones.service';
import { ToastService } from '../../../services/toast.service';
import type { Ocupacion, OcupacionInput } from '../../../models/ocupacion';
import { formatLocalInput, toLimaIso } from './ocupacion-utils';

@Component({
  selector: 'app-ocupacion-modal',
  templateUrl: './modal.html',
  styleUrl: './modal.css',
  imports: [FormsModule],
})
export class OcupacionModalComponent {
  private service = inject(OcupacionesService);
  private toast = inject(ToastService);

  ambienteId = input.required<number>();
  ocupacion = input<Ocupacion | null>(null);
  onClose = output<void>();
  onSaved = output<void>();

  isEditing = signal(false);
  saving = signal(false);
  error = signal('');

  form: { inicio: string; fin: string; activo: boolean } = {
    inicio: '',
    fin: '',
    activo: true,
  };

  constructor() {
    effect(() => {
      const ocupacion = this.ocupacion();
      if (ocupacion) {
        this.isEditing.set(true);
        this.form = {
          inicio: formatLocalInput(ocupacion.inicio),
          fin: formatLocalInput(ocupacion.fin),
          activo: ocupacion.activo,
        };
      } else {
        this.isEditing.set(false);
        this.form = { inicio: '', fin: '', activo: true };
      }
      this.error.set('');
    });
  }

  close(): void {
    this.onClose.emit();
  }

  onSubmit(): void {
    if (!this.form.inicio || !this.form.fin) {
      this.error.set('Complete la fecha y hora de inicio y fin');
      return;
    }
    if (new Date(this.form.inicio).getTime() >= new Date(this.form.fin).getTime()) {
      this.error.set('La fecha de inicio debe ser menor que la de fin');
      return;
    }

    this.saving.set(true);
    this.error.set('');

    const payload: OcupacionInput = {
      ambiente_id: this.ambienteId(),
      inicio: toLimaIso(this.form.inicio),
      fin: toLimaIso(this.form.fin),
      activo: this.form.activo,
    };

    const request = this.isEditing()
      ? this.service.update(this.ocupacion()!.id, payload)
      : this.service.create(payload);

    request.subscribe({
      next: () => {
        this.toast.success(this.isEditing() ? 'Ocupación actualizada' : 'Ocupación creada');
        this.saving.set(false);
        this.onSaved.emit();
        this.close();
      },
      error: (error) => {
        this.saving.set(false);
        this.error.set(
          error?.status === 409
            ? 'El horario se solapa con otra ocupación'
            : error?.status === 400
              ? 'Rango de fechas inválido'
              : 'Error al guardar la ocupación',
        );
      },
    });
  }
}
