import { environment } from '../../environments/environment';

export interface Imagen {
  id: number;
  ambiente_id: number;
  archivo: string;
  descripcion: string | null;
  orden: number;
  mime_type: string | null;
  bytes: number | null;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export function imagenUrl(imagen: Imagen): string {
  return `${environment.MEDIA_URL}/imagenes/${imagen.archivo}`;
}
