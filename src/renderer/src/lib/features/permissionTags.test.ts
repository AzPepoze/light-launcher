import { describe, expect, it } from "vitest";
import { permissionTag } from "./permissionTags";

describe("permissionTag", () => {
	it("marks sign-in details and file access as sensitive", () => {
		expect(permissionTag("secrets")).toEqual({ kind: "sign-in", sensitive: true });
		expect(permissionTag("fs:$GAMES")).toEqual({ kind: "files", sensitive: true });
	});

	it("keeps browser, program and site access normal", () => {
		expect(permissionTag("browser")).toEqual({ kind: "browser", sensitive: false });
		expect(permissionTag("exec")).toEqual({ kind: "program", sensitive: false });
		expect(permissionTag("network:example.com")).toEqual({ kind: "site", sensitive: false });
	});

	it("treats unknown entries as normal other access", () => {
		expect(permissionTag("camera")).toEqual({ kind: "other", sensitive: false });
		expect(permissionTag("constructor")).toEqual({ kind: "other", sensitive: false });
	});
});
