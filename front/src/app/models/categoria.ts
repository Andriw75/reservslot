export interface Categoria {
  id: number;
  nombre: string;
  descripcion: string;
  color: string | null;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CategoriaInput {
  nombre: string;
  descripcion: string;
  color?: string | null;
  activo?: boolean;
}
