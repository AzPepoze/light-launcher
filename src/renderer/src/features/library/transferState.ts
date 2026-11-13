import type { DownloadItem, InstallProgress } from "../../lib/features/types";

export type TransferKind = "downloading" | "installing";

/** The transfer running for each game id. An install of a game wins over its download. */
export function transferKinds(
	items: DownloadItem[],
	installs: InstallProgress[]
): Map<number, TransferKind> {
	const kinds = new Map<number, TransferKind>();
	for (const item of items) {
		if (item.gameId !== null && item.status === "downloading") {
			kinds.set(item.gameId, "downloading");
		}
	}
	for (const install of installs) {
		if (install.state === "active") kinds.set(install.gameId, "installing");
	}
	return kinds;
}
