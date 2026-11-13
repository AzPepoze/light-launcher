import type {
	Game,
	LibraryListPage,
	LibraryPageRequest,
	LibraryQuery
} from "../../lib/features/types";
import { PAGE_SIZE } from "./libraryWindow";
import { PageCache } from "./pageCache";

export const MAX_CACHED_PAGES = 20;

export type ListPage = (query: LibraryQuery, page: LibraryPageRequest) => Promise<LibraryListPage>;

export interface PagerSnapshot {
	/** Games that match the query, or null until the first page arrives. */
	total: number | null;
	/** Changes on every reset. Requests from an older generation are ignored. */
	generation: number;
	/** The last page error, or null when the last load worked. */
	error: unknown;
}

export interface PagerOptions {
	query: LibraryQuery;
	listPage: ListPage;
	onChange: (snapshot: PagerSnapshot) => void;
}

function sameQuery(a: LibraryQuery, b: LibraryQuery): boolean {
	return JSON.stringify(a) === JSON.stringify(b);
}

/**
 * Loads list pages on demand for one query and keeps the recent ones.
 * A query change or reset drops every page and ignores answers to older requests.
 */
export class LibraryPager {
	readonly pages = new PageCache<Game[]>(MAX_CACHED_PAGES);
	private query: LibraryQuery;
	private generation = 0;
	private total: number | null = null;
	private error: unknown = null;
	private readonly inflight = new Set<number>();
	private readonly listPage: ListPage;
	private readonly onChange: (snapshot: PagerSnapshot) => void;

	constructor(options: PagerOptions) {
		this.query = options.query;
		this.listPage = options.listPage;
		this.onChange = options.onChange;
	}

	/** Switches to another query. Does nothing when the query is the same. */
	setQuery(query: LibraryQuery): void {
		if (sameQuery(this.query, query)) return;
		this.query = query;
		this.reset();
	}

	/** Drops every page and loads the first one, which carries the total the grid needs to appear. */
	reset(): void {
		this.generation++;
		this.pages.clear();
		this.inflight.clear();
		this.total = null;
		this.error = null;
		this.notify();
		this.load(0);
	}

	/** Loads the pages that are not in the cache yet, and marks the cached ones as used. */
	requestPages(pages: number[], generation: number): void {
		if (generation !== this.generation) return;
		for (const page of pages) {
			if (this.pages.has(page)) {
				this.pages.touch(page);
			} else if (!this.inflight.has(page)) {
				this.load(page);
			}
		}
	}

	gameAt(index: number): Game | undefined {
		return this.pages.get(Math.floor(index / PAGE_SIZE))?.[index % PAGE_SIZE];
	}

	/** Replaces a game in every cached page that holds it. */
	replaceGame(game: Game): void {
		for (const [page, items] of this.pages.entries()) {
			if (items.some((item) => item.id === game.id)) {
				this.pages.set(
					page,
					items.map((item) => (item.id === game.id ? game : item))
				);
			}
		}
	}

	hasGame(id: number): boolean {
		return this.pages.entries().some(([, items]) => items.some((item) => item.id === id));
	}

	private load(page: number): void {
		const generation = this.generation;
		this.inflight.add(page);
		this.listPage(this.query, {
			offset: page * PAGE_SIZE,
			limit: PAGE_SIZE,
			wantTotal: page === 0
		})
			.then((result) => this.store(page, generation, result))
			.catch((err: unknown) => this.fail(generation, err))
			.finally(() => {
				if (generation === this.generation) this.inflight.delete(page);
			});
	}

	private store(page: number, generation: number, result: LibraryListPage): void {
		if (generation !== this.generation) return;
		if (page === 0) this.total = result.total ?? result.items.length;
		this.error = null;
		this.pages.set(page, result.items);
		this.notify();
	}

	private fail(generation: number, err: unknown): void {
		if (generation !== this.generation) return;
		this.error = err;
		this.notify();
	}

	private notify(): void {
		this.onChange({ total: this.total, generation: this.generation, error: this.error });
	}
}
