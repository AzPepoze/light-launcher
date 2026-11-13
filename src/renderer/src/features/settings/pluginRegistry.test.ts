import { describe, expect, it } from "vitest";
import { normalizePluginRegistryUrl } from "./pluginRegistry";

describe("normalizePluginRegistryUrl", () => {
	it("trims spaces around the URL", () => {
		expect(normalizePluginRegistryUrl("  https://plugins.example/plugins.json  ")).toBe(
			"https://plugins.example/plugins.json"
		);
	});

	it("keeps the URL itself unchanged", () => {
		expect(normalizePluginRegistryUrl("https://plugins.example/plugins.json")).toBe(
			"https://plugins.example/plugins.json"
		);
	});

	it("turns an empty or blank value into the disabled setting", () => {
		expect(normalizePluginRegistryUrl("")).toBe("");
		expect(normalizePluginRegistryUrl("   ")).toBe("");
		expect(normalizePluginRegistryUrl(undefined)).toBe("");
		expect(normalizePluginRegistryUrl(null)).toBe("");
	});

	it("leaves a non-https URL for the backend to reject", () => {
		expect(normalizePluginRegistryUrl(" http://plugins.example/plugins.json ")).toBe(
			"http://plugins.example/plugins.json"
		);
	});
});
