import type { GameFile } from "../../lib/features/types";

export type FilePlatform = "linux" | "windows" | "mac" | "android" | "other";

export interface FileGroup {
	platform: FilePlatform;
	label: string;
	icon: string;
	files: GameFile[];
}

const DEFAULT_ORDER: FilePlatform[] = ["linux", "windows", "mac", "android", "other"];

const PLATFORM_LABELS: Record<FilePlatform, string> = {
	linux: "Linux",
	windows: "Windows",
	mac: "macOS",
	android: "Android",
	other: "Other"
};

const PLATFORM_ICONS: Record<FilePlatform, string> = {
	linux: "terminal",
	windows: "desktop_windows",
	mac: "laptop_mac",
	android: "phone_android",
	other: "deployed_code"
};

/** Names that a file platform or an OS name can take, mapped to the platform group. */
const PLATFORM_BY_NAME: Record<string, FilePlatform> = {
	linux: "linux",
	windows: "windows",
	win: "windows",
	mac: "mac",
	macos: "mac",
	osx: "mac",
	darwin: "mac",
	android: "android"
};

/** Hosts that name the source, not the file server, so they add nothing to a row. */
const SOURCE_HOSTS = ["api.itch.io"];

const ARCHIVE_EXTENSIONS = ["zip", "7z", "rar", "tar", "gz", "tgz", "xz", "bz2"];
const INSTALLER_EXTENSIONS = ["exe", "msi", "appimage", "deb", "rpm", "dmg", "pkg"];
const WEB_EXTENSIONS = ["html", "htm"];

/** The platform group of a platform name or an OS name. Unknown names fall into Other. */
export function platformOf(name: string): FilePlatform {
	return PLATFORM_BY_NAME[name.trim().toLowerCase()] ?? "other";
}

export function platformLabel(platform: FilePlatform): string {
	return PLATFORM_LABELS[platform];
}

export function platformIcon(platform: FilePlatform): string {
	return PLATFORM_ICONS[platform];
}

/** The group order: the person's own platform first, then the rest in the usual order. */
export function platformOrder(ownPlatform: FilePlatform | null): FilePlatform[] {
	if (!ownPlatform) return DEFAULT_ORDER;
	return [ownPlatform, ...DEFAULT_ORDER.filter((platform) => platform !== ownPlatform)];
}

/** Groups the files by platform, in platform order. Platforms without files are left out. */
export function groupFiles(files: GameFile[], os: string): FileGroup[] {
	const ownPlatform = os ? platformOf(os) : null;
	const byPlatform = new Map<FilePlatform, GameFile[]>();
	for (const file of files) {
		const platform = platformOf(file.platform);
		byPlatform.set(platform, [...(byPlatform.get(platform) ?? []), file]);
	}
	return platformOrder(ownPlatform).flatMap((platform) => {
		const list = byPlatform.get(platform) ?? [];
		if (list.length === 0) return [];
		return [{ platform, label: platformLabel(platform), icon: platformIcon(platform), files: list }];
	});
}

/** The lower-case extension of a file name, or an empty string when it has none. */
export function extensionOf(name: string): string {
	const match = /\.([a-z0-9]+)$/i.exec(name.trim());
	return match ? match[1].toLowerCase() : "";
}

/** A Material Symbols icon for what the file is: an archive, an installer, an APK, a web page, or a plain file. */
export function fileIcon(file: GameFile): string {
	const extension = extensionOf(file.filename ?? file.label);
	if (extension === "apk") return "android";
	if (ARCHIVE_EXTENSIONS.includes(extension)) return "folder_zip";
	if (INSTALLER_EXTENSIONS.includes(extension)) return "install_desktop";
	if (WEB_EXTENSIONS.includes(extension)) return "web";
	return "draft";
}

/** The badge for a file kind the source names. Null for any other kind. */
export function kindBadge(kind: string): string | null {
	if (kind === "full") return "Full";
	if (kind === "patch") return "Patch";
	return null;
}

/** The host to show on a row, or null when it names the source rather than a file server. */
export function visibleHost(host: string | null): string | null {
	if (!host || SOURCE_HOSTS.includes(host)) return null;
	return host;
}
