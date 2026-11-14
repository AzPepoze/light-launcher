import { beforeEach, describe, expect, it, vi } from "vitest";
import { GetExeIcon, GetImageBase64 } from "@lib/api";
import { hasLibraryCover, loadGameIcon, needsExeIcon } from "./iconService";

vi.mock("@lib/api", () => ({
	GetExeIcon: vi.fn(),
	GetImageBase64: vi.fn()
}));

const exePath = "C:/Games/game.exe";

describe("needsExeIcon", () => {
	it("is false when the game has a remote or cached cover", () => {
		expect(needsExeIcon({ cover: "https://example.test/cover.jpg" })).toBe(false);
		expect(needsExeIcon({ coverLocal: "covers/1.jpg" })).toBe(false);
	});

	it("is true when the game has no cover", () => {
		expect(needsExeIcon({ cover: null, coverLocal: null })).toBe(true);
		expect(needsExeIcon({})).toBe(true);
	});
});

describe("hasLibraryCover", () => {
	it("reads both cover fields", () => {
		expect(hasLibraryCover({ cover: "https://example.test/cover.jpg" })).toBe(true);
		expect(hasLibraryCover({ coverLocal: "covers/1.jpg" })).toBe(true);
		expect(hasLibraryCover({ cover: "", coverLocal: null })).toBe(false);
	});
});

describe("loadGameIcon", () => {
	beforeEach(() => {
		vi.mocked(GetExeIcon).mockReset();
		vi.mocked(GetImageBase64).mockReset();
	});

	it("never calls GetExeIcon for a game with a cover", async () => {
		vi.mocked(GetExeIcon).mockResolvedValue("data:image/png;base64,exe");

		const result = await loadGameIcon(exePath, null, "low", true);

		expect(GetExeIcon).not.toHaveBeenCalled();
		expect(result).toEqual({ icon: "", source: null });
	});

	it("does not fall back to GetExeIcon when the custom icon fails for a game with a cover", async () => {
		vi.mocked(GetImageBase64).mockRejectedValue(new Error("missing"));
		vi.mocked(GetExeIcon).mockResolvedValue("data:image/png;base64,exe");

		const result = await loadGameIcon(exePath, "C:/icons/custom.png", "high", true);

		expect(GetExeIcon).not.toHaveBeenCalled();
		expect(result.source).toBeNull();
	});

	it("uses the custom icon first when it loads", async () => {
		vi.mocked(GetImageBase64).mockResolvedValue("data:image/png;base64,custom");
		vi.mocked(GetExeIcon).mockResolvedValue("data:image/png;base64,exe");

		const result = await loadGameIcon(exePath, "C:/icons/custom.png", "high", true);

		expect(result).toEqual({ icon: "data:image/png;base64,custom", source: "custom" });
		expect(GetExeIcon).not.toHaveBeenCalled();
	});

	it("extracts the exe icon for a game without a cover", async () => {
		vi.mocked(GetExeIcon).mockResolvedValue("data:image/png;base64,exe");

		const result = await loadGameIcon(exePath, null, "high", false);

		expect(GetExeIcon).toHaveBeenCalledWith(exePath);
		expect(result).toEqual({ icon: "data:image/png;base64,exe", source: "exe" });
	});
});
