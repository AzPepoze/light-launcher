import { RunGame } from "@lib/api";
import { notifications } from "@stores/notificationStore";
import type { LaunchOptions } from "@shared";

export interface LaunchRequest {
	showLogs?: boolean;
	announce?: boolean;
}

export async function launchGame(
	options: LaunchOptions,
	request: LaunchRequest = {}
): Promise<void> {
	const showLogs = request.showLogs ?? false;
	const announce = request.announce ?? true;
	const launchOptions = structuredClone(options);
	const name = launchOptions.Name || "game";

	if (announce) notifications.add(`Launching ${name}...`, "info");

	try {
		await RunGame(launchOptions, showLogs);
	} catch (error: unknown) {
		const message = error instanceof Error ? error.message : String(error);
		notifications.add(`Launch failed: ${message}`, "error");
		throw error;
	}
}
