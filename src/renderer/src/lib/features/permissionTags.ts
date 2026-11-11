import type { PermissionKind } from "./types";

/** Consent entry prefixes, as in core/internal/plugins/consent.go. */
export const NETWORK_PREFIX = "network:";
export const FILES_PREFIX = "fs:";

const FLAG_KINDS = new Map<string, PermissionKind>([
	["secrets", "sign-in"],
	["browser", "browser"],
	["exec", "program"]
]);

const SENSITIVE_KINDS: PermissionKind[] = ["sign-in", "files"];

export interface PermissionTag {
	kind: PermissionKind;
	/** Sign-in details and file access. The consent screen marks these in red. */
	sensitive: boolean;
}

export function permissionKind(entry: string): PermissionKind {
	if (entry.startsWith(NETWORK_PREFIX)) return "site";
	if (entry.startsWith(FILES_PREFIX)) return "files";
	return FLAG_KINDS.get(entry) ?? "other";
}

/** What kind of access a consent entry gives, and whether it is sensitive. */
export function permissionTag(entry: string): PermissionTag {
	const kind = permissionKind(entry);
	return { kind, sensitive: SENSITIVE_KINDS.includes(kind) };
}
