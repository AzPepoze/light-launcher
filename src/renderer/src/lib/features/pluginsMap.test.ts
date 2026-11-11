import { describe, expect, it } from "vitest";
import {
	consentLabel,
	isRemovedPlugin,
	newPermissionLines,
	permissionSummary,
	toLogLine,
	toPluginInfo,
	toPluginPreview,
	toPluginSettings,
	toPluginUpdatesStatus
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
describe("consentLabel", () => {
	it("writes a plain line for each kind of entry", () => {
		expect(consentLabel("network:api.example.test")).toBe("Connect to api.example.test");
		expect(consentLabel("fs:$GAMES")).toBe("Access files: $GAMES");
		expect(consentLabel("secrets")).toBe("Store sign-in details");
		expect(consentLabel("browser")).toBe("Open a browser window");
		expect(consentLabel("exec")).toBe("Run programs on this computer");
	});

	it("shows an unknown entry as it is", () => {
		expect(consentLabel("camera")).toBe("camera");
		expect(consentLabel("constructor")).toBe("constructor");
	});
});

describe("permissionSummary", () => {
	it("lists nothing when no permission is asked", () => {
		expect(permissionSummary(NO_PERMISSIONS)).toEqual([]);
	});

	it("lists every permission as a sorted consent entry", () => {
		expect(
			permissionSummary({
				network: ["b.example.test", "a.example.test"],
				fs: ["data"],
				secrets: true,
				browser: true,
				exec: true
			})
		).toEqual([
			"browser",
			"exec",
			"fs:data",
			"network:a.example.test",
			"network:b.example.test",
			"secrets"
		]);
	});
});

describe("newPermissionLines", () => {
	it("returns labels for the entries the installed consent does not have", () => {
		expect(
			newPermissionLines(
				["network:a.example.test"],
				["network:a.example.test", "network:b.example.test", "exec"]
			)
		).toEqual(["Connect to b.example.test", "Run programs on this computer"]);
	});

	it("returns nothing when the new version asks for nothing more", () => {
		expect(newPermissionLines(["exec"], ["exec"])).toEqual([]);
		expect(newPermissionLines(["exec", "browser"], ["browser"])).toEqual([]);
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
			repository: null,
			hasSettings: false,
			sources: ["example"],
			source: null,
			consent: []
		});
	});

	it("keeps the install source, repository and installed permissions", () => {
		const info = toPluginInfo(
			makeWire({
				source: "https://example.test/store.zip",
				repository: "github:ada/example",
				permissions: { ...NO_PERMISSIONS, exec: true, network: ["api.example.test"] }
			})
		);
		expect(info.source).toBe("https://example.test/store.zip");
		expect(info.repository).toBe("github:ada/example");
		expect(info.consent).toEqual(["exec", "network:api.example.test"]);
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
			consent: ["network:api.example.test"],
			repositoryChanged: false,
			installedRepository: null
		});
	});

	it("flags a package that names another repository", () => {
		const preview = toPluginPreview("https://example.test/p.zip", {
			manifest: {
				id: "com.example.store",
				name: "Example Store",
				version: "2.0.0",
				permissions: NO_PERMISSIONS
			},
			unsigned: false,
			consent: [],
			repositoryChanged: true,
			installedRepository: "github:ada/example"
		});
		expect(preview.repositoryChanged).toBe(true);
		expect(preview.installedRepository).toBe("github:ada/example");
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

describe("toPluginUpdatesStatus", () => {
	it("maps the cached check results and keeps plain text notes", () => {
		const status = toPluginUpdatesStatus({
			checkedAt: "2026-03-01T10:00:00Z",
			plugins: [
				{
					id: "com.example.store",
					installedVersion: "1.0.0",
					available: true,
					version: "1.1.0",
					assetUrl: "https://github.test/store.llplugin",
					sha256: "abc",
					notes: "<b>Fixes</b> a bug.",
					publishedAt: "2026-02-01T10:00:00Z"
				}
			]
		});
		expect(status.checkedAt).toBe("2026-03-01T10:00:00Z");
		expect(status.plugins[0]).toEqual({
			id: "com.example.store",
			installedVersion: "1.0.0",
			available: true,
			version: "1.1.0",
			assetUrl: "https://github.test/store.llplugin",
			sha256Url: null,
			sha256: "abc",
			notes: "<b>Fixes</b> a bug.",
			publishedAt: "2026-02-01T10:00:00Z",
			error: null
		});
	});

	it("returns an empty list when nothing was checked", () => {
		expect(toPluginUpdatesStatus(null)).toEqual({ checkedAt: null, plugins: [] });
		expect(toPluginUpdatesStatus({ plugins: null })).toEqual({ checkedAt: null, plugins: [] });
	});
});

describe("isRemovedPlugin", () => {
	it("tells a removal event from a state event", () => {
		expect(isRemovedPlugin({ id: "a", removed: true })).toBe(true);
		expect(isRemovedPlugin(makeWire())).toBe(false);
	});
});
