import { derived, get, writable } from "svelte/store";
import { GetPlatform } from "@lib/api";
import type { PlatformFeatures, PlatformInfo } from "@shared";
import { getRoute } from "../app/routes";

const noFeatures: PlatformFeatures = {
	proton: false,
	prefix: false,
	lsfg: false,
	gamescope: false,
	mangoHud: false,
	gameMode: false,
	liveTerminalLogs: false
};

/** Empty until loadPlatform() runs, so Linux-only parts stay hidden at first. */
export const platform = writable<PlatformInfo>({ os: "", features: noFeatures });

export const isLinux = derived(platform, ($platform) => $platform.os === "linux");

export async function loadPlatform(): Promise<void> {
	platform.set(await GetPlatform());
}

export function currentFeatures(): PlatformFeatures {
	return get(platform).features;
}

export function isPageAvailable(page: string, features: PlatformFeatures): boolean {
	const feature = getRoute(page)?.feature;
	return feature === undefined || features[feature];
}
