import { describe, it, expect } from "vitest";
import { ProtonService } from "./proton.service";
import type { ProtonTool } from "../../shared/types/prefix.types";

describe("ProtonService", () => {
	it("returns known proton variants", () => {
		const variants = ProtonService.getProtonVariants();
		expect(variants.length).toBeGreaterThanOrEqual(4);
		expect(variants.some((v) => v.ID === "ge-proton")).toBe(true);
		expect(variants.some((v) => v.ID === "proton-cachyos")).toBe(true);
		expect(variants.some((v) => v.ID === "kron4ek")).toBe(true);
		expect(variants.some((v) => v.ID === "luxtorpeda")).toBe(true);
	});

	it("matches proton tool by DisplayName", () => {
		const tools: ProtonTool[] = [
			{
				Name: "GE-Proton9-25",
				Path: "/home/user/.steam/root/compatibilitytools.d/GE-Proton9-25",
				IsSteam: false,
				DisplayName: "Custom: GE-Proton9-25"
			},
			{
				Name: "Proton 9.0",
				Path: "/home/user/.steam/root/steamapps/common/Proton 9.0",
				IsSteam: true,
				DisplayName: "Steam: Proton 9.0"
			}
		];

		const match1 = ProtonService.findProtonMatch("Custom: GE-Proton9-25", tools);
		expect(match1).toBeDefined();
		expect(match1?.Name).toBe("GE-Proton9-25");

		const match2 = ProtonService.findProtonMatch(
			"/home/user/.steam/root/steamapps/common/Proton 9.0",
			tools
		);
		expect(match2).toBeDefined();
		expect(match2?.IsSteam).toBe(true);

		const match3 = ProtonService.findProtonMatch("NonExistent", tools);
		expect(match3).toBeNull();
	});
});
