import { describe, expect, it } from "vitest";
import type { GameFile } from "../../lib/features/types";
import { queueGameFile } from "./gameFileQueue";

const file: GameFile = {
	id: 9,
	label: "Windows build",
	url: "https://files.example/win.zip",
	filename: null,
	size: null,
	host: null,
	kind: "direct",
	platform: "windows",
	requiresAuth: false
};

describe("queueGameFile", () => {
	it("marks the file busy while the host prepares it, then clears it", async () => {
		const events: string[] = [];
		const enqueue = async () => {
			events.push("enqueue");
			return 12;
		};

		const message = await queueGameFile(3, file, enqueue, (id) => events.push(`busy ${id}`));

		expect(events).toEqual(["busy 9", "enqueue", "busy null"]);
		expect(message).toBe('Added "Windows build" to downloads.');
	});

	it("sends the game id and the file id to the enqueue call", async () => {
		const calls: Array<[number, number]> = [];
		await queueGameFile(
			3,
			file,
			async (gameId, fileId) => {
				calls.push([gameId, fileId]);
				return 12;
			},
			() => {}
		);

		expect(calls).toEqual([[3, 9]]);
	});

	it("shows the plugin error and still clears the busy state", async () => {
		const states: Array<number | null> = [];
		const enqueue = async () => {
			throw new Error("the window stopped without a result");
		};

		const message = await queueGameFile(3, file, enqueue, (id) => states.push(id));

		expect(message).toBe("the window stopped without a result");
		expect(states).toEqual([9, null]);
	});

	it("shows a generic message for errors without text", async () => {
		const enqueue = async () => {
			throw "boom";
		};

		await expect(queueGameFile(3, file, enqueue, () => {})).resolves.toBe("Something went wrong.");
	});
});
