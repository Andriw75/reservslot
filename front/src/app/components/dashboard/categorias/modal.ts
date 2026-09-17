import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CategoriasService } from '../../../services/categorias.service';
import { ToastService } from '../../../services/toast.service';
import type { Categoria, CategoriaInput } from '../../../models/categoria';

@Component({
  selector: 'app-categoria-modal',
  templateUrl: './modal.html',
  styleUrl: './modal.css',
  imports: [FormsModule],
})
export class CategoriaModalComponent {
  private service = inject(CategoriasService);
  private toast = inject(ToastService);

  categoria = input<Categoria | null>(null);
  onClose = output<void>();
  onSaved = output<void>();

  isEditing = signal(false);
  saving = signal(false);
  error = signal('');

  form: CategoriaInput = {
    nombre: '',
    descripcion: '',
    color: null,
    activo: true,
  };

  constructor() {
    effect(() => {
      const categoria = this.categoria();
      if (categoria) {
        this.isEditing.set(true);
        this.form = {
          nombre: categoria.nombre,
          descripcion: categoria.descripcion,
          color: categoria.color,
          activo: categoria.activo,
        };
      } else {
        this.isEditing.set(false);
        this.form = { nombre: '', descripcion: '', color: null, activo: true };
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

    this.saving.set(true);
    this.error.set('');

    const payload: CategoriaInput = {
      nombre: this.form.nombre.trim(),
      descripcion: this.form.descripcion ?? '',
      color: this.form.color || null,
      activo: this.form.activo,
    };

    const request = this.isEditing()
      ? this.service.update(this.categoria()!.id, payload)
      : this.service.create(payload);

    request.subscribe({
      next: () => {
        this.toast.success(this.isEditing() ? 'Categoría actualizada' : 'Categoría creada');
        this.saving.set(false);
        this.onSaved.emit();
        this.close();
      },
      error: () => {
        this.saving.set(false);
        this.error.set('Error al guardar la categoría');
      },
    });
  }
}
