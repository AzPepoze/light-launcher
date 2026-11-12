import { launcherKindLabel } from "@lib/features/launcher";
import type { Launcher } from "@shared";

export interface LauncherMenuChoice {
	/** The picker value, or null when automatic selection is chosen. */
	path: string | null;
	icon: string;
	label: string;
	detail: string;
}

/** Menu rows: automatic first, then every launcher in the order the backend ranks them. */
export function launcherChoices(launchers: Launcher[]): LauncherMenuChoice[] {
	return [
		{ path: null, icon: "autorenew", label: "Use automatic", detail: "" },
		...launchers.map((launcher) => ({
			path: launcher.path,
			icon: launcher.primary ? "check" : "play_arrow",
			label: launcher.name,
			detail: launcherKindLabel(launcher.kind)
		}))
	];
}
