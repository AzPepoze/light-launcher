import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const root = fileURLToPath(new URL(".", import.meta.url));

/** The one raw `.btn` left: an external link needs href/target, which Button cannot carry. */
const LINK_EXCEPTION = "features/library/GameDetail.svelte";

function svelteFiles(dir: string): string[] {
	const files: string[] = [];
	for (const name of readdirSync(dir)) {
		const path = join(dir, name);
		if (statSync(path).isDirectory()) files.push(...svelteFiles(path));
		else if (name.endsWith(".svelte")) files.push(path);
	}
	return files;
}

describe("shared UI primitives", () => {
	it("uses Icon instead of material-icons spans", () => {
		const offenders = svelteFiles(root).filter((path) =>
			readFileSync(path, "utf8").includes('class="material-icons')
		);
		expect(offenders.map((path) => relative(root, path))).toEqual([]);
	});

	it("uses Button/IconButton instead of raw .btn classes", () => {
		const offenders = svelteFiles(root).filter((path) => {
			const name = relative(root, path);
			if (name === LINK_EXCEPTION) return false;
			return readFileSync(path, "utf8").includes('class="btn');
		});
		expect(offenders.map((path) => relative(root, path))).toEqual([]);
	});

	it("keeps the external-link exception confined to an anchor", () => {
		const path = join(root, LINK_EXCEPTION);
		const line = readFileSync(path, "utf8")
			.split("\n")
			.find((text) => text.includes('class="btn'));
		expect(line).toContain("<a ");
		expect(line).toContain("href=");
	});
});
