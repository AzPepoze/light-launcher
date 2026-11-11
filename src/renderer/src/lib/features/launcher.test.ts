import { describe, expect, it } from "vitest";
import { launcherCountLabel, launcherKindLabel } from "./launcher";

describe("launcherKindLabel", () => {
	it("names each launcher kind", () => {
		expect(launcherKindLabel("main")).toBe("Main game");
		expect(launcherKindLabel("unreal-stub")).toBe("Unreal launcher");
		expect(launcherKindLabel("x86")).toBe("32-bit");
		expect(launcherKindLabel("other")).toBe("Other");
	});

	it("falls back to Other for an unknown kind", () => {
		expect(launcherKindLabel("new-kind" as never)).toBe("Other");
	});
});

describe("launcherCountLabel", () => {
	it("counts launchers", () => {
		expect(launcherCountLabel(3)).toBe("3 launchers");
	});
});
