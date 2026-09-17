import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import type { Ambiente, AmbienteInput } from '../models/ambiente';
import type { AmbienteInfo, AmbienteInfoInput } from '../models/ambiente-info';
import type { Imagen } from '../models/imagen';

@Injectable({ providedIn: 'root' })
export class AmbientesService {
  private readonly api = environment.API_URL;
  private http = inject(HttpClient);

  list(): Observable<Ambiente[]> {
    return this.http.get<Ambiente[]>(`${this.api}/ambientes`);
  }

  getByID(id: number): Observable<Ambiente> {
    return this.http.get<Ambiente>(`${this.api}/ambientes/${id}`);
  }

  create(input: AmbienteInput): Observable<Ambiente> {
    return this.http.post<Ambiente>(`${this.api}/ambientes`, input);
  }

  update(id: number, input: Partial<AmbienteInput>): Observable<Ambiente> {
    return this.http.put<Ambiente>(`${this.api}/ambientes/${id}`, input);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/ambientes/${id}`);
  }

  getInfo(id: number): Observable<AmbienteInfo> {
    return this.http.get<AmbienteInfo>(`${this.api}/ambientes/${id}/info`);
  }

  updateInfo(id: number, input: AmbienteInfoInput): Observable<AmbienteInfo> {
    return this.http.put<AmbienteInfo>(`${this.api}/ambientes/${id}/info`, input);
  }

  uploadImagenes(id: number, files: File[]): Observable<Imagen[]> {
    const formData = new FormData();
    for (const file of files) {
      formData.append('images', file, file.name);
    }
    return this.http.post<Imagen[]>(`${this.api}/ambientes/${id}/imagenes`, formData);
  }

  deleteImagen(id: number, imagenId: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/ambientes/${id}/imagenes/${imagenId}`);
  }
}
