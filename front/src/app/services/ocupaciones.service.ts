import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import type { Ocupacion, OcupacionInput } from '../models/ocupacion';

export interface OcupacionesRange {
  from: string;
  to: string;
  data: Ocupacion[];
}

@Injectable({ providedIn: 'root' })
export class OcupacionesService {
  private readonly api = environment.API_URL;
  private http = inject(HttpClient);

  list(ambienteId: number, from?: string, to?: string): Observable<OcupacionesRange> {
    let params = new HttpParams();
    if (from && to) {
      params = params.set('from', from).set('to', to);
    }
    return this.http.get<OcupacionesRange>(`${this.api}/ambientes/${ambienteId}/ocupaciones`, {
      params,
    });
  }

  create(input: OcupacionInput): Observable<Ocupacion> {
    return this.http.post<Ocupacion>(`${this.api}/ocupaciones`, input);
  }

  update(id: number, input: Partial<OcupacionInput>): Observable<Ocupacion> {
    return this.http.put<Ocupacion>(`${this.api}/ocupaciones/${id}`, input);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/ocupaciones/${id}`);
  }
}
