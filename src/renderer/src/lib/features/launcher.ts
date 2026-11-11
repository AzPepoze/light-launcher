import type { LauncherKind } from "@shared";

const kindLabels: Record<LauncherKind, string> = {
	main: "Main game",
	"unreal-stub": "Unreal launcher",
	x86: "32-bit",
	other: "Other"
};

/** Short label for a launcher kind, shown under the launcher name. */
export function launcherKindLabel(kind: LauncherKind): string {
	return kindLabels[kind] ?? kindLabels.other;
}

/** Note shown on a card that has more than one launcher. */
export function launcherCountLabel(count: number): string {
	return `${count} launchers`;
}
