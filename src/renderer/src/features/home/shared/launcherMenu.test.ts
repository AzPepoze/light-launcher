import { describe, expect, it } from "vitest";
import type { Launcher } from "@shared";
import { launcherChoices } from "./launcherMenu";

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
