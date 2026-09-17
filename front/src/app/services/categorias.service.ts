import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import type { Categoria, CategoriaInput } from '../models/categoria';

@Injectable({ providedIn: 'root' })
export class CategoriasService {
  private readonly api = environment.API_URL;
  private http = inject(HttpClient);

  list(): Observable<Categoria[]> {
    return this.http.get<Categoria[]>(`${this.api}/categorias`);
  }

  getByID(id: number): Observable<Categoria> {
    return this.http.get<Categoria>(`${this.api}/categorias/${id}`);
  }

  create(input: CategoriaInput): Observable<Categoria> {
    return this.http.post<Categoria>(`${this.api}/categorias`, input);
  }

  update(id: number, input: Partial<CategoriaInput>): Observable<Categoria> {
    return this.http.put<Categoria>(`${this.api}/categorias/${id}`, input);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/categorias/${id}`);
  }
}
