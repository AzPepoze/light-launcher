/**
 * Shares one in-flight async load per key.
 *
 * A caller that asks for a key that is already loading receives the same
 * promise instead of an already-resolved one. That keeps "still loading"
 * distinct from "finished without a result" for every waiter (for example a
 * card that asks for an icon while the low-priority warm pass is loading it).
 */
export class IconLoadSharing<T = void> {
	private inFlight = new Map<string, Promise<T>>();

	share(key: string, start: () => Promise<T>): Promise<T> {
		const existing = this.inFlight.get(key);
		if (existing) return existing;

		const load = Promise.resolve()
			.then(start)
			.then(
				(value) => {
					this.inFlight.delete(key);
					return value;
				},
				(error: unknown) => {
					this.inFlight.delete(key);
					throw error;
				}
			);
		this.inFlight.set(key, load);
		return load;
	}
}
