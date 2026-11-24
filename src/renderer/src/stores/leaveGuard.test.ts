import { get } from "svelte/store";
import { afterEach, describe, expect, it, vi } from "vitest";
import { confirmLeave, leaveGuard, leavePrompt, type LeaveGuard } from "./leaveGuard";

function guardWith(dirty: boolean, save = vi.fn(async () => true)): LeaveGuard {
	return { isDirty: () => dirty, save };
}

/** Answers the open prompt as soon as it appears. */
function answerNextPrompt(choice: "stay" | "discard" | "save") {
	const unsubscribe = leavePrompt.subscribe((prompt) => {
		if (!prompt) return;
		unsubscribe();
		prompt.answer(choice);
	});
}

afterEach(() => {
	leaveGuard.set(null);
	leavePrompt.set(null);
});

describe("confirmLeave", () => {
	it("lets the user leave when no page holds a guard", async () => {
		await expect(confirmLeave()).resolves.toBe(true);
		expect(get(leavePrompt)).toBeNull();
	});

	it("lets the user leave a clean page without asking", async () => {
		leaveGuard.set(guardWith(false));
		await expect(confirmLeave()).resolves.toBe(true);
		expect(get(leavePrompt)).toBeNull();
	});

	it("keeps the page open when the user stays", async () => {
		const save = vi.fn(async () => true);
		leaveGuard.set(guardWith(true, save));
		answerNextPrompt("stay");

		await expect(confirmLeave()).resolves.toBe(false);
		expect(save).not.toHaveBeenCalled();
	});

	it("leaves without saving when the user discards", async () => {
		const save = vi.fn(async () => true);
		leaveGuard.set(guardWith(true, save));
		answerNextPrompt("discard");

		await expect(confirmLeave()).resolves.toBe(true);
		expect(save).not.toHaveBeenCalled();
	});

	it("saves and then leaves when the user saves", async () => {
		const save = vi.fn(async () => true);
		leaveGuard.set(guardWith(true, save));
		answerNextPrompt("save");

		await expect(confirmLeave()).resolves.toBe(true);
		expect(save).toHaveBeenCalledOnce();
	});

	it("keeps the page open when the save fails", async () => {
		leaveGuard.set(guardWith(true, vi.fn(async () => false)));
		answerNextPrompt("save");

		await expect(confirmLeave()).resolves.toBe(false);
	});

	it("closes the prompt once it is answered", async () => {
		leaveGuard.set(guardWith(true));
		answerNextPrompt("discard");
		await confirmLeave();

		expect(get(leavePrompt)).toBeNull();
	});
});
