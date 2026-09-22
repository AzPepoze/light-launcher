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
 * Bare names without an extension are treated as prefix matches, so "UnityCrashHandler"
 * behaves like "UnityCrashHandler*".
 */
export function compileWildcardPatterns(patterns: string[]): RegExp[] {
	const regexes: RegExp[] = [];
	for (const raw of patterns) {
		const pattern = raw?.trim();
		if (!pattern) continue;
		const effective = !pattern.includes("*") && !pattern.includes(".") ? `${pattern}*` : pattern;
		const regexStr = effective.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*");
		try {
			regexes.push(new RegExp(`^${regexStr}$`, "i"));
		} catch {
			// Skip invalid regex expressions safely
		}
	}
	return regexes;
}
