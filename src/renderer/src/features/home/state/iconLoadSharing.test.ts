import { describe, expect, it, vi } from "vitest";
import { IconLoadSharing } from "./iconLoadSharing";

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (error: unknown) => void;
	const promise = new Promise<T>((res, rej) => {
		resolve = res;
		reject = rej;
	});
	return { promise, resolve, reject };
}

describe("IconLoadSharing", () => {
	it("returns the same pending promise to every caller", async () => {
		const loads = new IconLoadSharing<string>();
		const gate = deferred<string>();
		const start = vi.fn(() => gate.promise);

		const first = loads.share("game.exe", start);
		const second = loads.share("game.exe", start);

		expect(second).toBe(first);
		await Promise.resolve();
		expect(start).toHaveBeenCalledTimes(1);

		gate.resolve("icon");
		await expect(first).resolves.toBe("icon");
	});

	it("keeps a late caller pending until the shared load finishes", async () => {
		const loads = new IconLoadSharing<string>();
		const gate = deferred<string>();
		const warm = loads.share("game.exe", () => gate.promise);

		let settled = false;
		const card = loads
			.share("game.exe", () => Promise.resolve("second load"))
			.then((value) => {
				settled = true;
				return value;
			});

		await Promise.resolve();
		expect(settled).toBe(false);

		gate.resolve("icon");
		await expect(card).resolves.toBe("icon");
		await expect(warm).resolves.toBe("icon");
	});

	it("shares a failure without caching it", async () => {
		const loads = new IconLoadSharing<string>();
		const gate = deferred<string>();
		const start = vi.fn(() => gate.promise);
		const first = loads.share("game.exe", start);
		const second = loads.share("game.exe", start);

		gate.reject(new Error("no icon"));
		await expect(first).rejects.toThrow("no icon");
		await expect(second).rejects.toThrow("no icon");

		start.mockResolvedValueOnce("icon");
		await expect(loads.share("game.exe", start)).resolves.toBe("icon");
		expect(start).toHaveBeenCalledTimes(2);
	});

	it("starts a new load after the previous one settles", async () => {
		const loads = new IconLoadSharing<string>();
		const start = vi.fn().mockResolvedValueOnce("first").mockResolvedValueOnce("second");

		await expect(loads.share("game.exe", start)).resolves.toBe("first");
		await expect(loads.share("game.exe", start)).resolves.toBe("second");
		expect(start).toHaveBeenCalledTimes(2);
	});

	it("keeps different keys independent", async () => {
		const loads = new IconLoadSharing<string>();
		const startA = vi.fn().mockResolvedValue("a");
		const startB = vi.fn().mockResolvedValue("b");

		const a = loads.share("a.exe", startA);
		const b = loads.share("b.exe", startB);

		expect(a).not.toBe(b);
		await expect(a).resolves.toBe("a");
		await expect(b).resolves.toBe("b");
	});
});
