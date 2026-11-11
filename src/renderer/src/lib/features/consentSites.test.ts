import { describe, expect, it } from "vitest";
import { siteLines } from "./consentSites";

describe("siteLines", () => {
	it("merges a wildcard and its bare host into one line", () => {
		expect(siteLines(["*.example.com", "example.com"])).toEqual([
			"example.com and its subdomains"
		]);
	});

	it("is the same whichever of the two comes first", () => {
		expect(siteLines(["example.com", "*.example.com"])).toEqual([
			"example.com and its subdomains"
		]);
	});

	it("describes a lone wildcard as any subdomain", () => {
		expect(siteLines(["*.mixdrop.co"])).toEqual(["any subdomain of mixdrop.co"]);
	});

	it("describes a lone bare host as itself", () => {
		expect(siteLines(["api.example.com"])).toEqual(["api.example.com"]);
	});

	it("keeps a wildcard and a bare host of different names apart", () => {
		expect(siteLines(["*.a.test", "b.test"])).toEqual(["any subdomain of a.test", "b.test"]);
	});

	it("collapses many near-duplicate hosts into one line per site", () => {
		const hosts = [
			"*.mixdrop.co",
			"mixdrop.co",
			"*.cdn.mixdrop.co",
			"cdn.mixdrop.co",
			"*.img.example.com",
			"img.example.com",
			"api.example.com"
		];

		expect(siteLines(hosts)).toEqual([
			"api.example.com",
			"cdn.mixdrop.co and its subdomains",
			"img.example.com and its subdomains",
			"mixdrop.co and its subdomains"
		]);
	});

	it("sorts the sites alphabetically by host", () => {
		expect(siteLines(["zeta.test", "alpha.test", "*.mid.test"])).toEqual([
			"alpha.test",
			"any subdomain of mid.test",
			"zeta.test"
		]);
	});

	it("keeps an uppercase host as it was given", () => {
		expect(siteLines(["API.Example.com", "*.API.Example.com"])).toEqual([
			"API.Example.com and its subdomains"
		]);
	});

	it("keeps an IDN host as it was given", () => {
		expect(siteLines(["*.bücher.de", "bücher.de", "xn--bcher-kva.de"])).toEqual([
			"bücher.de and its subdomains",
			"xn--bcher-kva.de"
		]);
	});

	it("returns nothing when there are no hosts", () => {
		expect(siteLines([])).toEqual([]);
	});
});
