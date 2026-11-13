import { beforeEach, describe, expect, it, vi } from "vitest";
import { communityEntries, pluginsApi } from "./pluginsApi";
import { rpc } from "./rpc";
import type {
	WirePermissions,
	WirePluginInfo,
	WireRegistryIndexEntry,
	WireTrustedKey
} from "./wire";
import type { PluginRegistryEntry } from "./types";

vi.mock("./rpc", () => ({ rpc: vi.fn() }));

const rpcMock = vi.mocked(rpc);

const NO_PERMISSIONS: WirePermissions = {
	network: null,
	fs: null,
	secrets: false,
	browser: false,
	exec: false
};

function wirePlugin(overrides: Partial<WirePluginInfo> = {}): WirePluginInfo {
	return {
		id: "demo",
		name: "Demo",
		version: "2.0.0",
		status: "stopped",
		enabled: true,
		running: false,
		idle: false,
		unsigned: false,
		permissions: NO_PERMISSIONS,
		sources: [],
		author: "Ada",
		description: "",
		homepage: "",
		hasSettings: false,
		commands: [],
		...overrides
	};
}

const CONSENT = ["network:example.test"];

describe("pluginsApi.installOrUpdate", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("sends the package and the consent list to PluginInstallOrUpdate", async () => {
		rpcMock.mockResolvedValue(wirePlugin());

		await pluginsApi.installOrUpdate("/uploads/demo.llplugin", true, CONSENT);

		expect(rpcMock).toHaveBeenCalledWith("PluginInstallOrUpdate", {
			source: "/uploads/demo.llplugin",
			allowUnsigned: true,
			acceptedPermissions: CONSENT
		});
	});

	it("returns the updated plugin the server sent", async () => {
		rpcMock.mockResolvedValue(wirePlugin({ version: "2.0.0", previous: "1.0.0" }));

		const info = await pluginsApi.installOrUpdate("/uploads/demo.llplugin", false, CONSENT);

		expect(info).toMatchObject({ id: "demo", version: "2.0.0", signed: true });
	});

	it("passes the same-version error on to the caller", async () => {
		rpcMock.mockRejectedValue(new Error("plugin is already at this version: demo 2.0.0"));

		await expect(
			pluginsApi.installOrUpdate("/uploads/demo.llplugin", true, CONSENT)
		).rejects.toThrow("plugin is already at this version: demo 2.0.0");
	});
});

describe("pluginsApi.install", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("stays a strict new install through PluginInstall", async () => {
		rpcMock.mockResolvedValue(wirePlugin());

		await pluginsApi.install("/uploads/demo.llplugin", true, CONSENT);

		expect(rpcMock).toHaveBeenCalledWith("PluginInstall", {
			source: "/uploads/demo.llplugin",
			allowUnsigned: true,
			acceptedPermissions: CONSENT
		});
	});
});

const TRUSTED_KEY: WireTrustedKey = {
	base64: "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
	fingerprint: "0123456789abcdef"
};

describe("pluginsApi trusted keys", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("lists trusted keys through PluginTrustedKeys", async () => {
		rpcMock.mockResolvedValue([TRUSTED_KEY]);

		const keys = await pluginsApi.trustedKeys();

		expect(rpcMock).toHaveBeenCalledWith("PluginTrustedKeys");
		expect(keys).toEqual([{ base64: TRUSTED_KEY.base64, fingerprint: TRUSTED_KEY.fingerprint }]);
	});

	it("treats a missing key list as empty", async () => {
		rpcMock.mockResolvedValue(null);

		await expect(pluginsApi.trustedKeys()).resolves.toEqual([]);
	});

	it("adds a key with PluginTrustedKeyAdd and returns the refreshed list", async () => {
		rpcMock.mockResolvedValue([TRUSTED_KEY]);

		const keys = await pluginsApi.addTrustedKey(TRUSTED_KEY.base64);

		expect(rpcMock).toHaveBeenCalledWith("PluginTrustedKeyAdd", { key: TRUSTED_KEY.base64 });
		expect(keys).toEqual([{ base64: TRUSTED_KEY.base64, fingerprint: TRUSTED_KEY.fingerprint }]);
	});

	it("removes a key with PluginTrustedKeyRemove and returns the refreshed list", async () => {
		rpcMock.mockResolvedValue([]);

		const keys = await pluginsApi.removeTrustedKey(TRUSTED_KEY.base64);

		expect(rpcMock).toHaveBeenCalledWith("PluginTrustedKeyRemove", { key: TRUSTED_KEY.base64 });
		expect(keys).toEqual([]);
	});
});

const REGISTRY_ENTRY: WireRegistryIndexEntry = {
	id: "demo",
	name: "Demo",
	description: "Finds demo games.",
	author: "Ada",
	repository: "github:ada/demo",
	version: "1.2.3",
	downloadUrl: "https://plugins.example/demo.llplugin",
	sha256: "a".repeat(64),
	homepage: "https://example.test/demo"
};

describe("pluginsApi.registry", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("maps the configured community index", async () => {
		rpcMock.mockResolvedValue({
			configured: true,
			url: "https://plugins.example/index.json",
			entries: [REGISTRY_ENTRY]
		});

		const registry = await pluginsApi.registry();

		expect(rpcMock).toHaveBeenCalledWith("PluginRegistry");
		expect(registry).toEqual({
			configured: true,
			url: "https://plugins.example/index.json",
			entries: [
				{
					id: "demo",
					name: "Demo",
					description: "Finds demo games.",
					author: "Ada",
					repository: "github:ada/demo",
					version: "1.2.3",
					downloadUrl: "https://plugins.example/demo.llplugin",
					sha256: "a".repeat(64),
					homepage: "https://example.test/demo"
				}
			],
			error: null
		});
	});

	it("keeps the unconfigured state and maps missing optional links to null", async () => {
		rpcMock.mockResolvedValue({
			configured: false,
			url: "",
			entries: [{ ...REGISTRY_ENTRY, repository: "", homepage: "" }]
		});

		const registry = await pluginsApi.registry();

		expect(registry.configured).toBe(false);
		expect(registry.url).toBe("");
		expect(registry.entries[0]).toMatchObject({ repository: null, homepage: null });
	});

	it("reads a load error from the result instead of failing the call", async () => {
		rpcMock.mockResolvedValue({
			configured: true,
			url: "https://plugins.example/index.json",
			entries: null,
			error: "could not fetch the community plugin index: HTTP 503"
		});

		const registry = await pluginsApi.registry();

		expect(registry.entries).toEqual([]);
		expect(registry.error).toBe("could not fetch the community plugin index: HTTP 503");
	});
});

describe("communityEntries", () => {
	function entry(overrides: Partial<PluginRegistryEntry> = {}): PluginRegistryEntry {
		return {
			id: "demo",
			name: "Demo",
			description: "",
			author: "Ada",
			repository: null,
			version: "1.2.3",
			downloadUrl: "https://plugins.example/demo.llplugin",
			sha256: "a".repeat(64),
			homepage: null,
			...overrides
		};
	}

	it("keeps installable entries that are not installed yet", () => {
		const kept = entry({ id: "kept" });
		const installed = entry({ id: "installed" });
		const httpOnly = entry({ id: "insecure", downloadUrl: "http://plugins.example/demo.llplugin" });
		const noDigest = entry({ id: "nodigest", sha256: "" });
		const shortDigest = entry({ id: "short", sha256: "abc" });

		expect(
			communityEntries([kept, installed, httpOnly, noDigest, shortDigest], ["installed"])
		).toEqual([kept]);
	});

	it("keeps the index order", () => {
		const first = entry({ id: "first" });
		const second = entry({ id: "second" });
		expect(communityEntries([first, second], [])).toEqual([first, second]);
	});
});
