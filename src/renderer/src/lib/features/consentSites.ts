const WILDCARD_PREFIX = "*.";

interface SiteKinds {
	wildcard: boolean;
	bare: boolean;
}

/** Hosts are grouped by the name they share, so "*.a.test" and "a.test" meet at "a.test". */
function baseHost(host: string): string {
	return host.startsWith(WILDCARD_PREFIX) ? host.slice(WILDCARD_PREFIX.length) : host;
}

function collectKinds(hosts: string[]): Map<string, SiteKinds> {
	const kinds = new Map<string, SiteKinds>();
	for (const host of hosts) {
		const base = baseHost(host);
		const entry = kinds.get(base) ?? { wildcard: false, bare: false };
		if (host.startsWith(WILDCARD_PREFIX)) entry.wildcard = true;
		else entry.bare = true;
		kinds.set(base, entry);
	}
	return kinds;
}

function siteText(base: string, kinds: SiteKinds): string {
	if (kinds.wildcard && kinds.bare) return `${base} and its subdomains`;
	if (kinds.wildcard) return `any subdomain of ${base}`;
	return base;
}

/**
 * One line per site, sorted by host. A wildcard and its bare host become one line.
 * Hosts keep the case and characters they were given.
 */
export function siteLines(hosts: string[]): string[] {
	const sorted = [...collectKinds(hosts)].sort(([a], [b]) => a.localeCompare(b, "en"));
	return sorted.map(([base, kinds]) => siteText(base, kinds));
}
