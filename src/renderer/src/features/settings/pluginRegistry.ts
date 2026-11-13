/**
 * Trims a community plugin index URL. An empty value disables the community list.
 * The URL scheme is not checked here: the backend rejects a value that is not https.
 */
export function normalizePluginRegistryUrl(raw: string | null | undefined): string {
	return (raw ?? "").trim();
}
