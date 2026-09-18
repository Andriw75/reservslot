import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AmbientesService } from '../../../services/ambientes.service';
import { ToastService } from '../../../services/toast.service';
import { ModalComponent } from '../../../common/ui/modal/modal';
import type { Ambiente, AmbienteInput } from '../../../models/ambiente';
import type { Categoria } from '../../../models/categoria';

@Component({
  selector: 'app-ambiente-modal',
  templateUrl: './modal.html',
  imports: [FormsModule, ModalComponent],
})
export class AmbienteModalComponent {
  private service = inject(AmbientesService);
  private toast = inject(ToastService);

  ambiente = input<Ambiente | null>(null);
  categorias = input<Categoria[]>([]);
  onClose = output<void>();
  onSaved = output<void>();

  isEditing = signal(false);
  saving = signal(false);
  error = signal('');

  form: { nombre: string; categoria_id: number | null; activo: boolean } = {
    nombre: '',
    categoria_id: null,
    activo: true,
  };

  constructor() {
    effect(() => {
      const ambiente = this.ambiente();
      const categorias = this.categorias();
      if (ambiente) {
        this.isEditing.set(true);
        this.form = {
          nombre: ambiente.nombre,
          categoria_id: ambiente.categoria_id,
          activo: ambiente.activo,
        };
      } else {
        this.isEditing.set(false);
        this.form = { nombre: '', categoria_id: categorias[0]?.id ?? null, activo: true };
      }
      this.error.set('');
    });
  }

  close(): void {
    this.onClose.emit();
  }

  onSubmit(): void {
    if (!this.form.nombre.trim()) {
      this.error.set('Nombre es obligatorio');
      return;
    }
    if (!this.form.categoria_id) {
      this.error.set('Seleccione una categoría');
      return;
    }

    this.saving.set(true);
    this.error.set('');

    const payload: AmbienteInput = {
      nombre: this.form.nombre.trim(),
      categoria_id: this.form.categoria_id,
      activo: this.form.activo,
    };

    const request = this.isEditing()
      ? this.service.update(this.ambiente()!.id, payload)
      : this.service.create(payload);

    request.subscribe({
      next: () => {
        this.toast.success(this.isEditing() ? 'Ambiente actualizado' : 'Ambiente creado');
        this.saving.set(false);
        this.onSaved.emit();
        this.close();
      },
      error: () => {
        this.saving.set(false);
        this.error.set('Error al guardar el ambiente');
      },
    });
  }
}
