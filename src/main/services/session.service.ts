import fs from "fs/promises";
import path from "path";
import { exec } from "child_process";
import { promisify } from "util";
import type { RunningSession } from "../../shared/types/games.types";

const execAsync = promisify(exec);

export class SessionService {
	static async getRunningSessions(): Promise<RunningSession[]> {
		const sessions: RunningSession[] = [];

		let pidsOutput = "";
		try {
			const { stdout } = await execAsync("pgrep light-launcher-instance");
			pidsOutput = stdout;
		} catch {
			try {
				const { stdout } = await execAsync("pgrep light-launcher-instan");
				pidsOutput = stdout;
			} catch {}
		}

		if (!pidsOutput.trim()) {
			return [];
		}

		const pids = pidsOutput
			.trim()
			.split("\n")
			.map((p) => parseInt(p.trim(), 10))
			.filter((p) => !isNaN(p) && p > 0);

		for (const pid of pids) {
			try {
				const cmdlinePath = `/proc/${pid}/cmdline`;
				const content = await fs.readFile(cmdlinePath, "utf-8");
				const args = content.split("\0");

				let gamePath = "";
				for (let i = 0; i < args.length; i++) {
					if (args[i] === "--game" && i + 1 < args.length) {
						gamePath = args[i + 1];
						break;
					}
				}

				if (gamePath) {
					const cleanPath = path.normalize(gamePath);
					const name = path.parse(path.basename(cleanPath)).name;
					sessions.push({
						pid,
						gamePath: cleanPath,
						gameName: name
					});
				}
			} catch {
				// Process might have terminated in between
			}
		}

		return sessions;
	}

	static async killSession(pid: number): Promise<void> {
		if (!pid || pid <= 0) return;
		try {
			process.kill(pid, "SIGINT");
		} catch (err: any) {
			if (err?.code !== "ESRCH") {
				console.error(`Failed to send SIGINT to pid ${pid}:`, err);
			}
		}
	}
}
