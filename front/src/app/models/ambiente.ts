export interface Ambiente {
  id: number;
  nombre: string;
  categoria_id: number;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface AmbienteInput {
  nombre: string;
  categoria_id: number;
  activo?: boolean;
}
