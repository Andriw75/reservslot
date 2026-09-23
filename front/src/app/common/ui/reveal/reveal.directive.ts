import { Directive, ElementRef, OnDestroy, OnInit, inject, input } from '@angular/core';

/**
 * Revela el elemento con una transición suave cuando entra en el viewport.
 * Respeta `prefers-reduced-motion` (ver `.reveal` en ui.css).
 *
 * Uso: <section appReveal [appRevealDelay]="120">…</section>
 */
@Directive({
  selector: '[appReveal]',
})
export class RevealDirective implements OnInit, OnDestroy {
  private element = inject<ElementRef<HTMLElement>>(ElementRef);
  private observer?: IntersectionObserver;

  delay = input(0, { alias: 'appRevealDelay' });

  ngOnInit(): void {
    const node = this.element.nativeElement;
    node.classList.add('reveal');

    const delay = this.delay();
    if (delay > 0) {
      node.style.setProperty('--reveal-delay', `${delay}ms`);
    }

    if (typeof IntersectionObserver === 'undefined') {
      node.classList.add('reveal--in');
      return;
    }

    this.observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            node.classList.add('reveal--in');
            this.observer?.disconnect();
          }
        }
      },
      { threshold: 0.12, rootMargin: '0px 0px -8% 0px' },
    );
    this.observer.observe(node);
  }

  ngOnDestroy(): void {
    this.observer?.disconnect();
  }
}
