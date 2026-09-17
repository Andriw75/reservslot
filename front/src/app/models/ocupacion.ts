export interface Ocupacion {
  id: number;
  ambiente_id: number;
  inicio: string;
  fin: string;
  activo: boolean;
  created_by?: number;
  updated_at: string;
}

export interface OcupacionInput {
  ambiente_id: number;
  inicio: string;
  fin: string;
  activo?: boolean;
}
