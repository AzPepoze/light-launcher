import { GetSystemToolsStatus, ScanProtonVersions } from "@lib/api";
import { notifications } from "@stores/notificationStore";
import type { LaunchOptions } from "@shared";
import {
	executeLaunch,
	getMissingTools,
	resolveLibraryProton,
	validateLaunchPrerequisites
} from "./runService";

export interface LaunchRequest {
	showLogs?: boolean;
	announce?: boolean;
	closeLauncher?: boolean;
}

/** Same pipeline as the Run page; missing tools surface as a toast. */
export async function launchGame(
	options: LaunchOptions,
	request: LaunchRequest = {}
): Promise<void> {
	const showLogs = request.showLogs ?? false;
	const announce = request.announce ?? true;
	const closeLauncher = request.closeLauncher ?? false;
	const launchOptions = structuredClone(options);
	const name = launchOptions.Name || "game";

	const prerequisiteError = validateLaunchPrerequisites(launchOptions);
	if (prerequisiteError) {
		notifications.add(prerequisiteError, "error");
		throw new Error(prerequisiteError);
	}

	const [systemStatus, protonVersions] = await Promise.all([
		GetSystemToolsStatus(),
		ScanProtonVersions()
	]);

	const missingTools = getMissingTools(launchOptions, systemStatus);
	if (missingTools.length > 0) {
		const message = `Missing tools: ${missingTools.join(", ")}. Install them or disable the integration.`;
		notifications.add(message, "error");
		throw new Error(message);
	}

	const selectedProton = await resolveLibraryProton(launchOptions, protonVersions || []);

	if (announce) notifications.add(`Launching ${name}...`, "info");

	await executeLaunch(launchOptions, selectedProton, protonVersions || [], showLogs, closeLauncher);
}
