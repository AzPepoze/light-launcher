import path from "path";

/**
 * Checks if targetPath is inside parentPath directory hierarchy.
 */
export function isSubPath(parentPath: string, targetPath: string): boolean {
	if (!parentPath || !targetPath) return false;
	const normalizedParent = path.normalize(parentPath);
	const normalizedTarget = path.normalize(targetPath);
	const rel = path.relative(normalizedParent, normalizedTarget);
	return !rel.startsWith("..") && !path.isAbsolute(rel) && rel !== "";
}

/**
 * Sanitizes and normalizes a file or folder path.
 */
export function cleanPath(targetPath: string): string {
	if (!targetPath) return "";
	return path.normalize(targetPath.trim());
}

/**
 * Converts glob-like wildcard string patterns (e.g. "*.exe", "unins*") to Case-Insensitive RegExps.
 */
export function compileWildcardPatterns(patterns: string[]): RegExp[] {
	const regexes: RegExp[] = [];
	for (const pattern of patterns) {
		if (!pattern) continue;
		const regexStr = pattern.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*");
		try {
			regexes.push(new RegExp(`^${regexStr}$`, "i"));
		} catch {
			// Skip invalid regex expressions safely
		}
	}
	return regexes;
}
