import { beforeEach, describe, expect, it, vi } from "vitest";
import { pluginsApi } from "./pluginsApi";
import { rpc } from "./rpc";
import type { WirePermissions, WirePluginInfo, WireTrustedKey } from "./wire";

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
