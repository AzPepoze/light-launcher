import { describe, expect, it } from "vitest";
import type { Launcher } from "@shared";
import {
	isRectVisible,
	launcherChoices,
	launcherMenuPosition,
	pickLauncherSide
} from "./launcherMenu";

const launchers: Launcher[] = [
	{ name: "Game", path: "/games/game.exe", kind: "main", primary: true },
	{ name: "Setup", path: "/games/setup.exe", kind: "x86", primary: false }
];

describe("launcherChoices", () => {
	it("offers automatic first, then each launcher", () => {
		expect(launcherChoices(launchers)).toEqual([
			{ path: null, icon: "autorenew", label: "Use automatic", detail: "" },
			{ path: "/games/game.exe", icon: "check", label: "Game", detail: "Main game" },
			{ path: "/games/setup.exe", icon: "play_arrow", label: "Setup", detail: "32-bit" }
		]);
	});

	it("still offers automatic when the folder has only one launcher", () => {
		expect(launcherChoices([launchers[0]]).map((choice) => choice.path)).toEqual([
			null,
			"/games/game.exe"
		]);
	});

	it("offers only automatic without launchers", () => {
		expect(launcherChoices([]).map((choice) => choice.path)).toEqual([null]);
	});
});

describe("pickLauncherSide", () => {
	it("keeps the pill on the right while the menu width fits to the window edge", () => {
		expect(pickLauncherSide(600, 1000)).toBe("right");
	});

	it("flips to the left when the card leaves less room than the menu on the right", () => {
		expect(pickLauncherSide(900, 1000)).toBe("left");
	});
});

describe("launcherMenuPosition", () => {
	const card = { left: 100, right: 300, top: 50 };

	it("opens 8px past the card's right edge on the right side", () => {
		expect(launcherMenuPosition(card, "right", 180)).toEqual({ x: 308, y: 50 });
	});

	it("opens so its right edge sits 8px before the card's left edge on the left side", () => {
		expect(launcherMenuPosition(card, "left", 180)).toEqual({ x: -88, y: 50 });
	});

	it("uses the real menu width on the left side", () => {
		expect(launcherMenuPosition(card, "left", 280)).toEqual({ x: -188, y: 50 });
	});

	it.each(["left", "right"] as const)("keeps the menu clear of the anchor on the %s side", (side) => {
		const anchor = { left: 100, right: 300, top: 50, bottom: 250 };
		const { x, y } = launcherMenuPosition(anchor, side, 280);
		const menu = { left: x, top: y, right: x + 280, bottom: y + 200 };
		expect(isRectVisible(menu, anchor)).toBe(false);
	});
});
