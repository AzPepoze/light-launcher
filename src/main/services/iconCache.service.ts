import fs from "fs/promises";
import os from "os";
import path from "path";
import crypto from "crypto";

export type IconCacheFormat = "png" | "ico";

/** Disk cache for extracted icons, keyed by exe path + mtime + size. */
export class IconCacheService {
	private static dir(): string {
		return path.join(os.tmpdir(), "light-launcher-icon-cache");
	}

	private static hash(executablePath: string, stat: { mtimeMs: number; size: number }): string {
		const key = `${executablePath}|${stat.mtimeMs}|${stat.size}|v2`;
		return crypto.createHash("sha1").update(key).digest("hex");
	}

	/** Returns a cached data URL, "" for a known icon-less exe, or null on miss. */
	static async get(
		executablePath: string,
		stat: { mtimeMs: number; size: number }
	): Promise<string | null> {
		const base = path.join(IconCacheService.dir(), IconCacheService.hash(executablePath, stat));
		for (const format of ["png", "ico"] as IconCacheFormat[]) {
			try {
				const bytes = await fs.readFile(`${base}.${format}`);
				if (!bytes.length) continue;
				const mime = format === "png" ? "image/png" : "image/x-icon";
				return `data:${mime};base64,${bytes.toString("base64")}`;
			} catch {}
		}
		try {
			await fs.access(`${base}.miss`);
			return "";
		} catch {
			return null;
		}
	}

	static async put(
		executablePath: string,
		stat: { mtimeMs: number; size: number },
		bytes: Buffer,
		format: IconCacheFormat
	): Promise<void> {
		if (!bytes.length) return;
		try {
			const dir = IconCacheService.dir();
			await fs.mkdir(dir, { recursive: true });
			const target = path.join(dir, `${IconCacheService.hash(executablePath, stat)}.${format}`);
			const tempFile = `${target}.${process.pid}.${Date.now()}.tmp`;
			await fs.writeFile(tempFile, bytes);
			await fs.rename(tempFile, target);
		} catch {}
	}

	/** Records that an exe has no extractable icon so extraction is not retried on later runs. */
	static async markMiss(
		executablePath: string,
		stat: { mtimeMs: number; size: number }
	): Promise<void> {
		try {
			const dir = IconCacheService.dir();
			await fs.mkdir(dir, { recursive: true });
			await fs.writeFile(path.join(dir, `${IconCacheService.hash(executablePath, stat)}.miss`), "");
		} catch {}
	}
}
