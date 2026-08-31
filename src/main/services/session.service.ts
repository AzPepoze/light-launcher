import fs from "fs/promises";
import path from "path";
import { exec } from "child_process";
import { promisify } from "util";
import { LoggerService } from "./logger.service";
import { ActivityService } from "./activity.service";
import type { RunningSession } from "../../shared/types/games.types";

const execAsync = promisify(exec);

export class SessionService {
	private static firstSeen = new Map<number, number>();

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
			this.firstSeen.clear();
			await ActivityService.syncSessions([]);
			return [];
		}

		const pids = pidsOutput
			.trim()
			.split("\n")
			.map((p) => parseInt(p.trim(), 10))
			.filter((p) => !isNaN(p) && p > 0);
		const activePids = new Set(pids);
		for (const pid of this.firstSeen.keys()) {
			if (!activePids.has(pid)) this.firstSeen.delete(pid);
		}

		for (const pid of pids) {
			try {
				const cmdlinePath = `/proc/${pid}/cmdline`;
				const content = await fs.readFile(cmdlinePath, "utf-8");
				const args = content.split("\0");

				let gamePath = "";
				let gameName = "";
				for (let i = 0; i < args.length; i++) {
					if (args[i] === "--game" && i + 1 < args.length) {
						gamePath = args[i + 1];
					}
					if (args[i] === "--game-name" && i + 1 < args.length) {
						gameName = args[i + 1];
					}
				}

				if (gamePath) {
					const cleanPath = path.normalize(gamePath);
					if (!gameName) gameName = path.parse(path.basename(cleanPath)).name;
					if (!this.firstSeen.has(pid)) this.firstSeen.set(pid, Date.now());
					sessions.push({
						pid,
						gamePath: cleanPath,
						gameName,
						startedAt: this.firstSeen.get(pid)
					});
				}
			} catch {
				// Process might have terminated in between
			}
		}

		await ActivityService.syncSessions(sessions);
		return sessions;
	}

	static async killSession(pid: number): Promise<void> {
		if (!pid || pid <= 0) return;
		try {
			process.kill(pid, "SIGINT");
		} catch (err: any) {
			if (err?.code !== "ESRCH") {
				LoggerService.error("Session", `Failed to send SIGINT to pid ${pid}: ${err}`);
			}
		}
	}
}
