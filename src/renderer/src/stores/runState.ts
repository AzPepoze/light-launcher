import { writable } from "svelte/store";
import * as core from "@shared";
import { DEFAULT_LAUNCH_OPTIONS } from "@lib/constants";

export interface RunState {
	mainExePath: string;
	gameIcon: string;
	launcherIcon: string;
	prefixPath: string;
	selectedPrefixName: string;
	selectedProton: string;
	options: core.LaunchOptions;
}

const initial: RunState = {
	mainExePath: "",
	gameIcon: "",
	launcherIcon: "",
	prefixPath: "",
	selectedPrefixName: "Default",
	selectedProton: "",
	options: DEFAULT_LAUNCH_OPTIONS,
};

export const runState = writable<RunState>(initial);
