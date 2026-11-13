import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import type { CommandPaletteState } from "./CommandPaletteState.svelte";
import type { WirePluginInfo } from "../../lib/features/wire";

// The stores and services below pull in the browser-only app graph, so the test
// replaces them with the smallest fakes the palette reads.
vi.mock("@stores/navigationStore", () => ({ navigationCommand: { set: vi.fn() } }));
vi.mock("@stores/notificationStore", () => ({
	notifications: { success: vi.fn(), error: vi.fn() }
}));
vi.mock("@stores/libraryTabStore", () => ({ openLibraryTab: vi.fn() }));
vi.mock("@stores/platformStore", () => ({
	currentFeatures: () => ({
		proton: true,
		prefix: true,
		lsfg: true,
		gamescope: true,
		mangoHud: true,
		gameMode: true,
		liveTerminalLogs: true
	}),
	isPageAvailable: () => true
}));
vi.mock("@lib/api", () => ({
	GetAllGames: vi.fn(async () => []),
	GetImageBase64: vi.fn(async () => "")
}));
vi.mock("@lib/iconService", () => ({ loadExeIcon: vi.fn(async () => "") }));
vi.mock("@lib/homeService", () => ({ quickLaunchGame: vi.fn() }));
// The palette only needs the page entries, not the image imports in the real routes file.
vi.mock("../../app/routes", () => ({ routes: [] }));
// Keep the real errorMessage and replace only the transport call.
vi.mock("../../lib/features/rpc", async (importOriginal) => {
	const actual = await importOriginal<typeof import("../../lib/features/rpc")>();
	return { ...actual, rpc: vi.fn() };
});

import { notifications } from "@stores/notificationStore";
import { rpc } from "../../lib/features/rpc";

const rpcMock = vi.mocked(rpc);
const notificationsMock = vi.mocked(notifications);

type StateCtor = typeof import("./CommandPaletteState.svelte").CommandPaletteState;

let CommandPaletteStateClass: StateCtor;
let state: CommandPaletteState;

/** Answers each RPC the palette makes, by method name. */
function answerWith(answers: Record<string, unknown>) {
	rpcMock.mockImplementation(async (method: string) => {
		if (!(method in answers)) throw new Error(`unexpected call ${method}`);
		return answers[method] as never;
	});
}

/** A plugin row as the server sends it. Only the fields the palette reads are filled in. */
function wirePlugin(overrides: Partial<WirePluginInfo> = {}): WirePluginInfo {
	return {
		id: "com.example.store",
		name: "Example Store",
		version: "1.0.0",
		status: "running",
		enabled: true,
		running: true,
		idle: false,
		unsigned: false,
		permissions: { network: null, fs: null, secrets: false, browser: false, exec: false },
		sources: [],
		author: "Ada",
		description: "Finds games.",
		homepage: "https://example.test",
		hasSettings: false,
		commands: null,
		...overrides
	};
}

beforeAll(async () => {
	// The suite has no Svelte compiler plugin, so the runes are plain calls here.
	// The palette only replaces whole state fields, which an identity rune keeps working.
	vi.stubGlobal("$state", <T>(value: T): T => value);
	vi.stubGlobal("HTMLElement", class {});
	vi.stubGlobal("document", { contains: () => false, activeElement: null });

	const module = await import("./CommandPaletteState.svelte");
	CommandPaletteStateClass = module.CommandPaletteState;
});

afterAll(() => {
	vi.unstubAllGlobals();
});

beforeEach(() => {
	rpcMock.mockReset();
	notificationsMock.success.mockReset();
	notificationsMock.error.mockReset();
	state = new CommandPaletteStateClass();
});

describe("plugin commands in the command palette", () => {
	it("lists a Plugin entry for each command and runs that command when picked", async () => {
		answerWith({
			PluginsList: [
				wirePlugin({
					commands: [
						{ id: "import-key", label: "Import key", method: "example.importKey" }
					]
				})
			],
			PluginCall: null
		});

		await state.loadPluginCommands();

		const entry = state.filteredItems.find((item) => item.type === "command");
		// The name carries the plugin and command labels; type "command" is what
		// CommandPalette.svelte renders as the "Plugin" badge.
		expect(entry).toEqual({
			name: "Example Store: Import key",
			icon: "extension",
			type: "command",
			pluginId: "com.example.store",
			commandId: "import-key"
		});

		await state.executeItem(entry);

		expect(rpcMock).toHaveBeenLastCalledWith("PluginCall", {
			plugin: "com.example.store",
			command: "import-key",
			params: {}
		});
		expect(notificationsMock.success).toHaveBeenCalledWith("Example Store: Import key finished.");
		expect(state.show).toBe(false);
	});

	it("adds no entries and does not throw for a plugin without commands", async () => {
		answerWith({ PluginsList: [wirePlugin({ commands: null })] });

		await expect(state.loadPluginCommands()).resolves.toBeUndefined();

		expect(state.pluginCommands).toEqual([]);
		expect(state.filteredItems.filter((item) => item.type === "command")).toEqual([]);
	});
});
