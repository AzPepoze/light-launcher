// OS names from the platform store ("linux", "windows", "darwin") and the platform names on game files.
const FILE_PLATFORM_BY_OS: Record<string, string> = {
	linux: "linux",
	windows: "windows",
	darwin: "mac"
};

/** True when a game file can run on this OS. Files with no platform, and an unknown OS, always match. */
export function isFileForOs(filePlatform: string, os: string): boolean {
	if (!os || !filePlatform || filePlatform === "unknown") return true;
	return FILE_PLATFORM_BY_OS[os] === filePlatform;
}
