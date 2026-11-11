import { describe, expect, it } from "vitest";
import {
	isRemovedPlugin,
	permissionLabels,
	toLogLine,
	toPluginInfo,
	toPluginPreview,
	toPluginSettings
} from "./pluginsMap";
import type { WirePermissions, WirePluginInfo } from "./wire";

const NO_PERMISSIONS: WirePermissions = {
	network: null,
	fs: null,
	secrets: false,
	browser: false,
	exec: false
};

function makeWire(overrides: Partial<WirePluginInfo> = {}): WirePluginInfo {
	return {
		id: "com.example.store",
		name: "Example Store",
		version: "1.0.0",
		status: "stopped",
		enabled: true,
		running: false,
		unsigned: false,
		permissions: NO_PERMISSIONS,
		sources: [
			{ plugin: "com.example.store", sourceId: "example", label: "Example", capabilities: [] }
		],
		author: "Ada",
		description: "Finds games.",
		homepage: "https://example.test",
		hasSettings: false,
		...overrides
	};
}

describe("permissionLabels", () => {
	it("lists nothing when no permission is asked", () => {
		expect(permissionLabels(NO_PERMISSIONS)).toEqual([]);
	});

	it("writes one line per network host and file path, then the flags", () => {
		expect(
			permissionLabels({
				network: ["api.example.test"],
				fs: ["data"],
				secrets: true,
				browser: true,
				exec: true
			})
		).toEqual([
			"Connect to api.example.test",
			"Access files: data",
			"Store sign-in details",
			"Open a browser window",
			"Run programs on this computer"
		]);
	});
});

describe("toPluginInfo", () => {
	it("maps the Go status to ok or error and reads source ids", () => {
		expect(toPluginInfo(makeWire())).toEqual({
			id: "com.example.store",
			name: "Example Store",
			version: "1.0.0",
			enabled: true,
			signed: true,
			status: "ok",
			error: null,
			disabledReason: null,
			author: "Ada",
			description: "Finds games.",
			homepage: "https://example.test",
			hasSettings: false,
			sources: ["example"]
		});
	});

	it("passes the disabled reason and the settings flag through", () => {
		const info = toPluginInfo(
			makeWire({ enabled: false, disabledReason: "crashed 5 times", hasSettings: true })
		);
		expect(info.disabledReason).toBe("crashed 5 times");
		expect(info.hasSettings).toBe(true);
	});

	it("shows a broken plugin as an error with its reason", () => {
		const info = toPluginInfo(makeWire({ status: "broken", error: "plugin.json missing" }));
		expect(info.status).toBe("error");
		expect(info.error).toBe("plugin.json missing");
	});

	it("marks unsigned plugins and treats no sources as an empty list", () => {
		const info = toPluginInfo(makeWire({ unsigned: true, sources: null }));
		expect(info.signed).toBe(false);
		expect(info.sources).toEqual([]);
	});
});

describe("toPluginPreview", () => {
	it("uses the source the user typed and the manifest details", () => {
		const preview = toPluginPreview("https://example.test/p.zip", {
			manifest: {
				id: "com.example.store",
				name: "Example Store",
				version: "1.0.0",
				permissions: { ...NO_PERMISSIONS, network: ["api.example.test"] }
			},
			unsigned: true,
			consent: ["network:api.example.test"]
		});
		expect(preview).toEqual({
			source: "https://example.test/p.zip",
			id: "com.example.store",
			name: "Example Store",
			version: "1.0.0",
			signed: false,
			permissions: ["Connect to api.example.test"],
			consent: ["network:api.example.test"]
		});
	});
});

describe("toLogLine", () => {
	it("maps the level, time and text the launcher sent", () => {
		expect(
			toLogLine({ time: "2026-03-01T10:00:00Z", level: "warn", text: "retrying login" })
		).toEqual({ time: "2026-03-01T10:00:00Z", level: "warn", message: "retrying login" });
	});

	it("counts a line with no or an unknown level as info", () => {
		expect(toLogLine({ time: "t", text: "plain" }).level).toBe("info");
		expect(toLogLine({ time: "t", level: "trace", text: "x" }).level).toBe("info");
	});
});

describe("toPluginSettings", () => {
	it("passes the schema from plugin.json and the saved values", () => {
		const schema = { type: "object" as const, properties: { region: { type: "string" as const } } };
		expect(toPluginSettings({ schema, values: { region: "eu" } })).toEqual({
			schema,
			values: { region: "eu" }
		});
	});

	it("has no schema and empty values for a plugin without settings", () => {
		expect(toPluginSettings({ schema: null, values: null })).toEqual({ schema: null, values: {} });
	});
});

describe("isRemovedPlugin", () => {
	it("tells a removal event from a state event", () => {
		expect(isRemovedPlugin({ id: "a", removed: true })).toBe(true);
		expect(isRemovedPlugin(makeWire())).toBe(false);
	});
});
