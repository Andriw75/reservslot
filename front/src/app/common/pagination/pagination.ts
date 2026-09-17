import { Component, input, output, signal, computed, HostListener } from '@angular/core';

type PageItem = number | 'ellipsis';

@Component({
  selector: 'app-pagination',
  templateUrl: './pagination.html',
  styleUrl: './pagination.css',
})
export class PaginationComponent {
  page = input.required<number>();
  totalPages = input.required<number>();
  total = input<number>(0);
  pageChange = output<number>();

  private range = 2;
  inputValue = signal('');
  isOpen = signal(false);

  allPages = computed(() => {
    const tp = this.totalPages();
    return Array.from({ length: tp }, (_, i) => i + 1);
  });

  filteredPages = computed(() => {
    const value = this.inputValue().trim();
    if (!value) return this.allPages();
    return this.allPages().filter((p) => p.toString().startsWith(value));
  });

  beforeItems = computed((): PageItem[] => {
    const items: PageItem[] = [];
    const current = this.page();
    const tp = this.totalPages();
    const r = this.range;
    if (!tp || current <= 1) return items;
    items.push(1);
    if (current - r > 2) items.push('ellipsis');
    for (let i = Math.max(2, current - r); i < current; i++) {
      items.push(i);
    }
    return items;
  });

  afterItems = computed((): PageItem[] => {
    const items: PageItem[] = [];
    const current = this.page();
    const tp = this.totalPages();
    const r = this.range;
    if (!tp || current >= tp) return items;
    for (let i = current + 1; i <= Math.min(tp - 1, current + r); i++) {
      items.push(i);
    }
    if (current + r < tp - 1) items.push('ellipsis');
    items.push(tp);
    return items;
  });

  onFocus(): void {
    this.inputValue.set(this.page().toString());
    this.isOpen.set(true);
  }

  onInput(event: Event): void {
    const el = event.target as HTMLInputElement;
    this.inputValue.set(el.value);
    this.isOpen.set(true);
  }

  onKeyDown(event: KeyboardEvent): void {
    if (event.key !== 'Enter') return;
    const p = Number(this.inputValue());
    if (isNaN(p) || p < 1 || p > this.totalPages() || p === this.page()) {
      this.isOpen.set(false);
      return;
    }
    this.pageChange.emit(p);
    this.isOpen.set(false);
  }

  goTo(p: number): void {
    if (p === this.page()) return;
    this.pageChange.emit(p);
  }

  selectPage(p: number): void {
    if (p === this.page()) return;
    this.pageChange.emit(p);
    this.isOpen.set(false);
  }

  jump(offset: number): void {
    const target = this.page() + offset;
    if (target >= 1 && target <= this.totalPages()) {
      this.pageChange.emit(target);
    }
  }

  @HostListener('document:click', ['$event'])
  onClickOutside(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    if (!target.closest('.combo-wrapper')) {
      this.isOpen.set(false);
    }
  }
}
