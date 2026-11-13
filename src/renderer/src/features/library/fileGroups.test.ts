import { describe, expect, it } from "vitest";
import type { GameFile } from "../../lib/features/types";
import {
	extensionOf,
	fileIcon,
	groupFiles,
	kindBadge,
	platformOf,
	platformOrder,
	visibleHost
} from "./fileGroups";

function file(id: number, platform: string, filename: string | null = null): GameFile {
	return {
		id,
		label: `File ${id}`,
		url: "https://example.test/file",
		filename,
		size: null,
		host: null,
		kind: "",
		platform,
		requiresAuth: false
	};
}

describe("platformOf", () => {
	it("maps platform and OS names to the platform groups", () => {
		expect(platformOf("linux")).toBe("linux");
		expect(platformOf("windows")).toBe("windows");
		expect(platformOf("darwin")).toBe("mac");
		expect(platformOf("mac")).toBe("mac");
		expect(platformOf("android")).toBe("android");
	});

	it("puts unknown and empty names in Other", () => {
		expect(platformOf("unknown")).toBe("other");
		expect(platformOf("")).toBe("other");
	});
});

describe("platformOrder", () => {
	it("puts the own platform first", () => {
		expect(platformOrder("windows")).toEqual(["windows", "linux", "mac", "android", "other"]);
	});

	it("keeps the usual order when the own platform is unknown", () => {
		expect(platformOrder(null)).toEqual(["linux", "windows", "mac", "android", "other"]);
	});
});

describe("groupFiles", () => {
	const files = [file(1, "mac"), file(2, "linux"), file(3, "unknown"), file(4, "linux")];

	it("groups by platform, the own OS first, and skips empty groups", () => {
		const groups = groupFiles(files, "linux");
		expect(groups.map((group) => group.platform)).toEqual(["linux", "mac", "other"]);
		expect(groups[0].files.map((item) => item.id)).toEqual([2, 4]);
	});

	it("gives each group its label and icon", () => {
		const [group] = groupFiles(files, "linux");
		expect(group.label).toBe("Linux");
		expect(group.icon).toBe("terminal");
	});

	it("uses the usual order when the OS is not known", () => {
		expect(groupFiles(files, "").map((group) => group.platform)).toEqual(["linux", "mac", "other"]);
	});
});

describe("extensionOf", () => {
	it("returns the lower-case extension", () => {
		expect(extensionOf("Setup.EXE")).toBe("exe");
	});

	it("returns an empty string without an extension", () => {
		expect(extensionOf("readme")).toBe("");
	});
});

describe("fileIcon", () => {
	it("uses the file name first, so the label does not decide the icon", () => {
		expect(fileIcon(file(1, "linux", "sample-game.zip"))).toBe("folder_zip");
		expect(fileIcon(file(2, "android", "sample-game.apk"))).toBe("android");
		expect(fileIcon(file(3, "windows", "setup.msi"))).toBe("install_desktop");
		expect(fileIcon(file(4, "linux", "index.html"))).toBe("web");
	});

	it("falls back to a plain file icon", () => {
		expect(fileIcon(file(5, "linux", "notes.txt"))).toBe("draft");
	});

	it("reads the label when there is no file name", () => {
		expect(fileIcon({ ...file(6, "linux"), label: "Sample.7z" })).toBe("folder_zip");
	});
});

describe("kindBadge", () => {
	it("names the full and patch kinds only", () => {
		expect(kindBadge("full")).toBe("Full");
		expect(kindBadge("patch")).toBe("Patch");
		expect(kindBadge("mirror")).toBeNull();
	});
});

describe("visibleHost", () => {
	it("hides the source host and empty hosts", () => {
		expect(visibleHost("api.itch.io")).toBeNull();
		expect(visibleHost(null)).toBeNull();
	});

	it("keeps other hosts", () => {
		expect(visibleHost("files.example.test")).toBe("files.example.test");
	});
});
