import { SvelteMap } from "svelte/reactivity";

/** Keeps the most recently used pages, up to a fixed count. Reads stay reactive. */
export class PageCache<T> {
	private readonly items = new SvelteMap<number, T>();
	private readonly capacity: number;

	constructor(capacity: number) {
		this.capacity = capacity;
	}

	get size(): number {
		return this.items.size;
	}

	has(page: number): boolean {
		return this.items.has(page);
	}

	get(page: number): T | undefined {
		return this.items.get(page);
	}

	/** Marks a page as used most recently. */
	touch(page: number): void {
		const value = this.items.get(page);
		if (value === undefined) return;
		this.items.delete(page);
		this.items.set(page, value);
	}

	/** Stores a page. When the cache is full, the least recently used page is dropped first. */
	set(page: number, value: T): void {
		if (!this.items.has(page)) this.makeRoom();
		this.items.set(page, value);
	}

	clear(): void {
		this.items.clear();
	}

	entries(): [number, T][] {
		return [...this.items.entries()];
	}

	private makeRoom(): void {
		while (this.items.size >= this.capacity) {
			const oldest: number | undefined = this.items.keys().next().value;
			if (oldest === undefined) return;
			this.items.delete(oldest);
		}
	}
}
