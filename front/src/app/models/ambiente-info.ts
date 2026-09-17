import type { Imagen } from './imagen';

export interface AtributoItem {
  clave: string;
  valor: string;
}

export interface AmbienteInfoData {
  ancho?: number | null;
  alto?: number | null;
  aforo?: number | null;
  descripcion?: string;
  atributos?: AtributoItem[];
  opciones_adicionales?: Record<string, boolean>;
  [key: string]: unknown;
}

export interface AmbienteInfo {
  ambiente_id: number;
  data: AmbienteInfoData;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface AmbienteInfoInput {
  data: AmbienteInfoData;
  activo?: boolean;
}

export interface AmbienteInfoResponse {
  info: AmbienteInfo;
  imagenes: Imagen[];
}
