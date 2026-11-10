import { describe, expect, it } from "vitest";
import {
	isRemovedPlugin,
	parseLogLine,
	permissionLabels,
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
			sources: ["example"]
		});
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
			unsigned: true
		});
		expect(preview).toEqual({
			source: "https://example.test/p.zip",
			id: "com.example.store",
			name: "Example Store",
			version: "1.0.0",
			signed: false,
			permissions: ["Connect to api.example.test"]
		});
	});
});

describe("parseLogLine", () => {
	it("reads the level from the bracket prefix", () => {
		expect(parseLogLine("[warn] retrying login")).toEqual({
			time: "",
			level: "warn",
			message: "retrying login"
		});
	});

	it("keeps a line with no known level as info", () => {
		expect(parseLogLine("plain text")).toEqual({ time: "", level: "info", message: "plain text" });
	});
});

describe("toPluginSettings", () => {
	it("has no schema and empty values when nothing is saved", () => {
		expect(toPluginSettings(null)).toEqual({ schema: null, values: {} });
		expect(toPluginSettings({ region: "eu" })).toEqual({ schema: null, values: { region: "eu" } });
	});
});

describe("isRemovedPlugin", () => {
	it("tells a removal event from a state event", () => {
		expect(isRemovedPlugin({ id: "a", removed: true })).toBe(true);
		expect(isRemovedPlugin(makeWire())).toBe(false);
	});
});
