import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { cleanPath, compileWildcardPatterns } from "./pathUtils";

export interface ScanOptions {
	maxDepth?: number;
	extensions?: string[];
	excludePatterns?: string[];
	onlyFiles?: boolean;
}

/**
 * Recursively scans a directory for files matching criteria with customizable depth and regex exclusions.
 */
export async function scanDirectoryFiles(
	dirPath: string,
	options: ScanOptions = {}
): Promise<string[]> {
	const cleanDir = cleanPath(dirPath);
	if (!cleanDir || !fsSync.existsSync(cleanDir)) {
		return [];
	}

	const maxDepth = options.maxDepth !== undefined ? options.maxDepth : -1;
	const excludeRegexes = options.excludePatterns
		? compileWildcardPatterns(options.excludePatterns)
		: [];
	const allowedExtensions = options.extensions
		? options.extensions.map((ext) => ext.toLowerCase())
		: null;

	const results: string[] = [];

	async function scan(currentDir: string, currentDepth: number): Promise<void> {
		if (maxDepth !== -1 && currentDepth > maxDepth) {
			return;
		}

		let entries: fsSync.Dirent[];
		try {
			entries = await fs.readdir(currentDir, { withFileTypes: true });
		} catch {
			return;
		}

		for (const entry of entries) {
			const name = entry.name;
			if (excludeRegexes.some((regex) => regex.test(name))) {
				continue;
			}

			const fullPath = path.join(currentDir, name);
			try {
				if (entry.isDirectory()) {
					await scan(fullPath, currentDepth + 1);
				} else if (entry.isFile()) {
					if (
						!allowedExtensions ||
						allowedExtensions.some((ext) => name.toLowerCase().endsWith(ext))
					) {
						results.push(fullPath);
					}
				}
			} catch {
				// Skip unreadable files or broken symlinks safely
			}
		}
	}

	await scan(cleanDir, 0);
	return results;
}
