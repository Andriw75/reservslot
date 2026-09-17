import { Injectable } from '@angular/core';
import { Observable, Subject } from 'rxjs';
import { environment } from '../../environments/environment';

export interface SseMessage {
  event: string;
  topic: string;
  data: any;
}

const EVENT_NAMES = [
  'snapshot',
  'categoria.upsert',
  'categoria.delete',
  'ambiente.upsert',
  'ambiente.delete',
  'ambiente_info.upsert',
  'ocupacion.upsert',
  'ocupacion.delete',
];

@Injectable({ providedIn: 'root' })
export class SseService {
  private eventSource: EventSource | null = null;
  private currentTopics = '';
  private readonly subject = new Subject<SseMessage>();

  readonly events$: Observable<SseMessage> = this.subject.asObservable();

  connect(topics: string[]): void {
    const unique = [...new Set(topics)].filter((topic) => topic.length > 0);
    const key = unique.join(',');
    if (this.eventSource && key === this.currentTopics) return;

    this.disconnect();
    this.currentTopics = key;
    this.eventSource = new EventSource(`${environment.API_URL}/events?topics=${key}`, {
      withCredentials: true,
    });
    for (const name of EVENT_NAMES) {
      this.eventSource.addEventListener(name, (event) => this.handle(name, event));
    }
    this.eventSource.onerror = () => {};
  }

  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.currentTopics = '';
  }

  private handle(name: string, event: Event): void {
    const raw = (event as MessageEvent).data;
    if (typeof raw !== 'string') return;

    let payload: { topic?: string; data?: unknown };
    try {
      payload = JSON.parse(raw);
    } catch {
      return;
    }
    this.subject.next({ event: name, topic: payload.topic ?? '', data: payload.data });
  }
}
