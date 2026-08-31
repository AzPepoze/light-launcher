import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { ConfigService } from "./config.service";
import { PathsService } from "./paths.service";
import { LoggerService } from "./logger.service";
import type { LaunchOptions } from "../../shared/types/config.types";
import type { GameActivity, RunningSession } from "../../shared/types/games.types";

interface ActivityStore {
	games: Record<string, GameActivity>;
}

export class ActivityService {
	private static getStorePath(): string {
		return path.join(PathsService.getBaseDirectory(), "activity.json");
	}

	private static key(gamePath: string): string {
		return path.normalize(gamePath).toLowerCase();
	}

	private static async loadStore(): Promise<ActivityStore> {
		const storePath = this.getStorePath();
		if (!fsSync.existsSync(storePath)) {
			return { games: {} };
		}

		try {
			const store = await ConfigService.loadJson<ActivityStore>(storePath);
			return store?.games ? store : { games: {} };
		} catch (err) {
			LoggerService.warn("Activity", `Failed to load activity store: ${err}`);
			return { games: {} };
		}
	}

	private static async saveStore(store: ActivityStore): Promise<void> {
		await fs.mkdir(path.dirname(this.getStorePath()), { recursive: true });
		await ConfigService.saveJson(this.getStorePath(), store);
	}

	static async recordLaunch(options: LaunchOptions): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		if (settings.TrackPlaytime === false) return;

		const now = Date.now();
		const gamePath = path.normalize(options.GamePath || options.LauncherPath);
		if (!gamePath) return;

		const store = await this.loadStore();
		const key = this.key(gamePath);
		const previous = store.games[key];

		store.games[key] = {
			gamePath,
			gameName: options.Name || path.parse(gamePath).name,
			profileId: options.ID || previous?.profileId,
			customIconPath: options.CustomIconPath || previous?.customIconPath,
			lastPlayedAt: now,
			totalPlaytimeSeconds: previous?.totalPlaytimeSeconds || 0,
			sessionCount: previous?.activeSince ? previous.sessionCount : (previous?.sessionCount || 0) + 1,
			activeSince: previous?.activeSince || now
		};

		await this.saveStore(store);
	}

	static async syncSessions(sessions: RunningSession[]): Promise<GameActivity[]> {
		const settings = await ConfigService.loadAppSettings();
		const store = await this.loadStore();
		const now = Date.now();
		let changed = false;

		if (settings.TrackPlaytime === false) {
			for (const activity of Object.values(store.games)) {
				if (activity.activeSince) {
					delete activity.activeSince;
					changed = true;
				}
			}
			if (changed) await this.saveStore(store);
			return this.sortActivities(store);
		}

		const running = new Map(sessions.map((session) => [this.key(session.gamePath), session]));

		// Playtime is finalized by the Go instance; renderer only clears stale markers.
		for (const [key, activity] of Object.entries(store.games)) {
			if (activity.activeSince && !running.has(key)) {
				delete activity.activeSince;
				changed = true;
			}
		}

		// Rehydrate active sessions after launcher restart.
		for (const session of sessions) {
			const key = this.key(session.gamePath);
			const existing = store.games[key];
			if (!existing) {
				store.games[key] = {
					gamePath: session.gamePath,
					gameName: session.gameName,
					lastPlayedAt: session.startedAt || now,
					totalPlaytimeSeconds: 0,
					sessionCount: 1,
					activeSince: session.startedAt || now
				};
				changed = true;
			} else if (!existing.activeSince) {
				existing.activeSince = session.startedAt || now;
				existing.lastPlayedAt = Math.max(existing.lastPlayedAt, session.startedAt || now);
				existing.sessionCount += 1;
				changed = true;
			}
		}

		if (changed) await this.saveStore(store);
		return this.sortActivities(store);
	}

	static async getActivities(): Promise<GameActivity[]> {
		return this.sortActivities(await this.loadStore());
	}

	static async getActivitiesWithSessionSync(): Promise<GameActivity[]> {
		const { SessionService } = await import("./session.service.js");
		await SessionService.getRunningSessions();
		return this.getActivities();
	}

	private static sortActivities(store: ActivityStore): GameActivity[] {
		return Object.values(store.games).sort((a, b) => b.lastPlayedAt - a.lastPlayedAt);
	}
}
