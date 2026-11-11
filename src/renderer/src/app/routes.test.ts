import { describe, expect, it } from "vitest";
import { getRoute, routes } from "./routes";

describe("route registry", () => {
	it("keeps ids unique and order in array sequence", () => {
		const ids = routes.map((route) => route.id);
		expect(new Set(ids).size).toBe(ids.length);
		routes.forEach((route, index) => expect(route.order).toBe(index));
	});

	it("gates the Linux-only pages on platform features", () => {
		expect(getRoute("versions")?.feature).toBe("proton");
		expect(getRoute("prefix")?.feature).toBe("prefix");
		expect(getRoute("editlsfg")?.feature).toBe("lsfg");
	});

	it("lists every page except the fullscreen editor in navigation", () => {
		const hidden = routes.filter((route) => route.hidden).map((route) => route.id);
		expect(hidden).toEqual(["editlsfg"]);
	});
});
