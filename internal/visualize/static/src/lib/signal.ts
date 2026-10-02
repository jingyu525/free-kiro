// signal.ts — minimal reactive primitive. We don't pull in nanostores
// or Preact-signals because the dashboard has <13 components and
// adding a dependency for ~60 lines of TS is a bad trade.
//
// Usage:
//
//   const counter = new Signal(0);
//   counter.subscribe((v) => console.log(v));
//   counter.set(counter.get() + 1);

export type Unsubscribe = () => void;
export type Listener<T> = (value: T) => void;

export class Signal<T> {
  private value: T;
  private listeners = new Set<Listener<T>>();

  constructor(initial: T) {
    this.value = initial;
  }

  get(): T {
    return this.value;
  }

  set(next: T): void {
    if (Object.is(next, this.value)) return;
    this.value = next;
    // Iterate over a snapshot so a listener that unsubscribes itself
    // during dispatch doesn't trip the Set iterator.
    for (const fn of [...this.listeners]) fn(next);
  }

  // update is sugar for read-modify-write. Throwing from `fn`
  // propagates to the caller; listeners never see a partial value.
  update(fn: (prev: T) => T): void {
    this.set(fn(this.value));
  }

  subscribe(fn: Listener<T>): Unsubscribe {
    this.listeners.add(fn);
    return () => {
      this.listeners.delete(fn);
    };
  }
}
