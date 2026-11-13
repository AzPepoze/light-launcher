import { beforeEach, describe, expect, it, vi } from "vitest";
import { pluginsApi } from "./pluginsApi";
import { rpc } from "./rpc";
import type { WirePermissions, WirePluginInfo } from "./wire";

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
