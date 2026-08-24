<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import StatusDrawer from "@components/shared/StatusDrawer.svelte";
	import HeatmapCard from "@components/manager/HeatmapCard.svelte";
	import RunningRail from "@components/manager/RunningRail.svelte";
	import ActivityTabs from "@components/manager/ActivityTabs.svelte";
	import { GetRunningSessions, KillSession, GetAppSettings, GetGameActivity, GetAllGames, GetAutoScannedGames } from "@lib/api";
	import { notifications } from "@stores/notificationStore";
	import type { AppSettings, GameActivity, RunningSession, GameInfo, ScannedFolderGroup } from "@shared";

	// ===== MOCK PREVIEW — set false for real data =====
	const USE_MOCK = true;
	const MIN = 60 * 1000;
	const HOUR = 60 * MIN;
	const DAY = 24 * HOUR;

	type DaySession = { gameName: string; seconds: number };
	type DaySessionMap = Record<string, DaySession[]>;

	function dayKeyOf(ts: number): string {
		const d = new Date(ts);
		d.setHours(0, 0, 0, 0);
		return d.toISOString().slice(0, 10);
	}

	function buildMockActivities(now: number): GameActivity[] {
		const rows: [string, string, number, number, number][] = [
			["Elden Ring", "/games/eldenring/eldenring.exe", 4523, 14, 0],
			["Hades", "/games/hades/hades.exe", 12500, 32, 1],
			["Cyberpunk 2077", "/games/cyberpunk/Cyberpunk2077.exe", 38200, 58, 2],
			["Baldur's Gate 3", "/games/bg3/bg3_dx11.exe", 51200, 41, 4],
			["Stardew Valley", "/games/stardew/Stardew Valley.exe", 22400, 87, 6],
			["Hollow Knight", "/games/hollow/hollow_knight.exe", 9800, 21, 9],
			["Celeste", "/games/celeste/Celeste.exe", 5400, 12, 13],
			["Doom Eternal", "/games/doom/DOOMEternalx64vk.exe", 14700, 19, 18],
			["Factorio", "/games/factorio/factorio.exe", 60100, 120, 24],
			["Sekiro", "/games/sekiro/sekiro.exe", 30100, 26, 33],
			["Portal 2", "/games/portal2/portal2.exe", 8200, 9, 47],
			["Terraria", "/games/terraria/terraria.exe", 16800, 44, 62],
			["The Witcher 3", "/games/witcher3/witcher3.exe", 41300, 37, 88],
			["Dark Souls III", "/games/ds3/DarkSoulsIII.exe", 27600, 29, 130],
			["Cuphead", "/games/cuphead/Cuphead.exe", 3900, 7, 168]
		];
		return rows.map(([name, path, total, count, daysAgo], i) => ({
			gamePath: path,
			gameName: name,
			lastPlayedAt: now - daysAgo * DAY - (i % 5) * HOUR,
			totalPlaytimeSeconds: total,
			sessionCount: count,
			...(daysAgo === 0 ? { activeSince: now - 14 * MIN } : {})
		})) as GameActivity[];
	}

	const MOCK_GAME_NAMES = [
		"Elden Ring", "Hades", "Cyberpunk 2077", "Baldur's Gate 3", "Stardew Valley",
		"Hollow Knight", "Celeste", "Doom Eternal", "Factorio", "Sekiro",
		"Portal 2", "Terraria", "The Witcher 3", "Dark Souls III", "Cuphead"
	];

	function buildMockDaySessions(now: number): DaySessionMap {
		// per-day per-game playtime so heatmap + tooltip have real data
		const today = new Date(now);
		today.setHours(0, 0, 0, 0);
		const map: DaySessionMap = {};
		for (let d = 181; d >= 1; d--) {
			const rnd = Math.abs(Math.sin(d * 12.9898) * 43758.5453) % 1;
			const dayDate = new Date(today.getTime() - d * DAY);
			const weekend = [0, 6].includes(dayDate.getDay());
			let gameCount = 0;
			if (rnd > 0.5) gameCount = 1;
			if (rnd > 0.75) gameCount = 2;
			if (rnd > 0.9) gameCount = 3;
			if (weekend && rnd > 0.4) gameCount += 1;
			if (gameCount === 0) continue;
			const sessions: DaySession[] = [];
			for (let g = 0; g < gameCount; g++) {
				const seed = Math.abs(Math.sin((d * 7 + g * 13) * 4.231) * 15678.21) % 1;
				const nameIdx = Math.floor(Math.abs(Math.sin((d + g * 31) * 7.777) * 9876.5) % 1 * MOCK_GAME_NAMES.length);
				sessions.push({
					gameName: MOCK_GAME_NAMES[nameIdx],
					seconds: Math.floor((20 * 60) + seed * (2.5 * 3600 - 20 * 60)) // 20m .. ~2.5h
				});
			}
			map[dayKeyOf(dayDate.getTime())] = sessions;
		}
		// today: matches the running mock sessions
		map[dayKeyOf(now)] = [
			{ gameName: "Elden Ring", seconds: 4523 },
			{ gameName: "Hades", seconds: 3120 }
		];
		return map;
	}

	// Real mode: activity.json only stores lastPlayedAt + lifetime total, so per-day
	// times fall back to the game's total playtime on its last-played day.
	function buildDaySessionsFromActivities(acts: GameActivity[]): DaySessionMap {
		const map: DaySessionMap = {};
		for (const a of acts) {
			const key = dayKeyOf(a.lastPlayedAt);
			if (!map[key]) map[key] = [];
			map[key].push({ gameName: a.gameName, seconds: a.totalPlaytimeSeconds });
		}
		return map;
	}

	const mockSessions = (now: number): RunningSession[] => [
		{ pid: 4242, gamePath: "/games/eldenring/eldenring.exe", gameName: "Elden Ring", startedAt: now - 14 * MIN },
		{ pid: 4243, gamePath: "/games/hades/hades.exe", gameName: "Hades", startedAt: now - 52 * MIN }
	];

	const mockGames: GameInfo[] = [
		{ name: "Slay the Spire", path: "/games/sts/SlayTheSpire.exe", icon: "", config: {} as any, isRecent: false, isAutoScanned: false },
		{ name: "RimWorld", path: "/games/rimworld/RimWorld.exe", icon: "", config: {} as any, isRecent: false, isAutoScanned: false }
	];
	// ===== end mock =====

	let sessions: RunningSession[] = [];
	let activities: GameActivity[] = [];
	let daySessions: DaySessionMap = {};
	let appSettings: AppSettings | null = null;
	let games: GameInfo[] = [];
	let scannedGroups: ScannedFolderGroup[] = [];
	let refreshTimer: ReturnType<typeof setInterval> | null = null;
	let clockTimer: ReturnType<typeof setInterval> | null = null;
	let now = Date.now();

	$: trackingEnabled = appSettings?.TrackPlaytime !== false;

	// today derived
	$: todayStart = (()=>{ const d=new Date(now); d.setHours(0,0,0,0); return d.getTime(); })();
	$: todayKey = dayKeyOf(todayStart);
	// persisted per-day time + live delta from running games launched today
	$: todayPersisted = (daySessions[todayKey] || []).reduce((s, g) => s + g.seconds, 0);
	$: liveDelta = activities.reduce((sum, a) => {
		if (!a.activeSince || a.activeSince < todayStart) return sum;
		return sum + Math.max(0, Math.floor((now - a.activeSince) / 1000));
	}, 0);
	$: todayPlaytimeSeconds = todayPersisted + liveDelta;
	$: sessionsToday = activities.filter(a=> a.lastPlayedAt >= todayStart).length;
	$: runningCount = sessions.length;

	// recent: last 7 days
	$: recent = activities.filter(a=> now - a.lastPlayedAt < 7*DAY).slice(0,12);
	$: recentDisplay = recent.length>0 ? recent : activities.slice(0,8);

	$: mostPlayed = [...activities].sort((a,b)=> b.totalPlaytimeSeconds - a.totalPlaytimeSeconds).slice(0,12);

	// stale: library games never played or >30d
	$: libraryPaths = (()=> {
		const set = new Map<string,{name:string,path:string}>();
		for(const g of games) set.set(g.path.toLowerCase(), {name:g.name, path:g.path});
		for(const grp of scannedGroups) for(const g of grp.games) set.set(g.path.toLowerCase(), {name:g.name, path:g.path});
		return set;
	})();
	$: stale = (()=> {
		const THIRTY_D = 30*DAY;
		const actMap = new Map<string, GameActivity>();
		for(const a of activities) actMap.set(a.gamePath.toLowerCase(), a);
		const out: GameActivity[] = [];
		for(const [key, g] of libraryPaths){
			const a = actMap.get(key);
			if(!a){
				out.push({ gamePath:g.path, gameName:g.name, lastPlayedAt:0, totalPlaytimeSeconds:0, sessionCount:0 } as GameActivity);
			} else if(!a.activeSince && now - a.lastPlayedAt > THIRTY_D){
				out.push(a);
			}
		}
		for(const a of activities){
			if(a.activeSince) continue;
			if(now - a.lastPlayedAt > THIRTY_D && !out.find(x=> x.gamePath.toLowerCase()===a.gamePath.toLowerCase())){
				if(!libraryPaths.has(a.gamePath.toLowerCase())) out.push(a);
			}
		}
		out.sort((a,b)=> (a.lastPlayedAt||0) - (b.lastPlayedAt||0));
		return out.slice(0,12);
	})();

	onMount(() => {
		void refresh();
		refreshTimer = setInterval(() => void refresh(), 3000);
		clockTimer = setInterval(() => (now = Date.now()), 1000);
	});

	onDestroy(() => {
		if (refreshTimer) clearInterval(refreshTimer);
		if (clockTimer) clearInterval(clockTimer);
	});

	async function refresh() {
		if (USE_MOCK) {
			activities = buildMockActivities(now);
			daySessions = buildMockDaySessions(now);
			sessions = mockSessions(now);
			games = mockGames;
			scannedGroups = [];
			appSettings = { TrackPlaytime: true } as AppSettings;
			return;
		}
		try {
			const [running, history, settings, allGames, groups] = await Promise.all([
				GetRunningSessions(),
				GetGameActivity(),
				GetAppSettings(),
				GetAllGames().catch(()=>[] as GameInfo[]),
				GetAutoScannedGames().catch(()=>[] as ScannedFolderGroup[])
			]);
			sessions = running || [];
			activities = history || [];
			daySessions = buildDaySessionsFromActivities(activities);
			appSettings = settings;
			games = allGames || [];
			scannedGroups = groups || [];
		} catch (err) {
			console.error("Failed to refresh game manager:", err);
		}
	}

	async function stopSession(session: RunningSession) {
		if (USE_MOCK) {
			notifications.add(`[mock] Stopping ${session.gameName}...`, "info");
			return;
		}
		try {
			await KillSession(session.pid);
			notifications.add(`Stopping ${session.gameName}...`, "info");
			setTimeout(() => void refresh(), 400);
		} catch (err) {
			notifications.add(`Failed to stop ${session.gameName}: ${err}`, "error");
		}
	}

	function formatTodayDuration(sec:number){
		const s=Math.max(0,Math.floor(sec));
		const h=Math.floor(s/3600), m=Math.floor((s%3600)/60);
		if(h>0) return `${h}h ${m}m`;
		if(m>0) return `${m}m`;
		return `${s}s`;
	}
</script>

<div class="manager-container">
	<!-- Top: single transparent centered container -->
	<div class="top-summary">
		<div class="today-block">
			<span class="today-label">Today</span>
			<span class="today-value">{formatTodayDuration(todayPlaytimeSeconds)}</span>
			<span class="today-sub">{sessionsToday} {sessionsToday===1?'session':'sessions'} · {runningCount} running</span>
		</div>
		<div class="heatmap-block">
			<HeatmapCard {daySessions} {now} />
		</div>
	</div>

	<div class="manager-grid">
		<div class="left-col">
			<div class="card">
				<RunningRail {sessions} {activities} {now} onStop={stopSession} />
			</div>
		</div>
		<div class="right-col">
			<div class="card">
				<ActivityTabs
					recent={recentDisplay}
					{mostPlayed}
					{stale}
					{now}
					{trackingEnabled}
				/>
			</div>
		</div>
	</div>

	<!-- System status as fixed bottom drawer (old main style) -->
	<StatusDrawer />
</div>

<style lang="scss">
	.manager-container {
		display: flex;
		flex-direction: column;
		gap: 26px;
		min-height: calc(100vh - 200px);
		padding-bottom: 140px; // room for fixed drawer
		box-sizing: border-box;
	}

	.top-summary {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 56px;
		flex-wrap: wrap;
	}

	.today-block {
		display: flex;
		flex-direction: column;
		gap: 4px;
		.today-label{ font-size:0.7rem; letter-spacing:1px; text-transform:uppercase; color:var(--text-dim); font-weight:800; }
		.today-value{ font-size:2rem; font-weight:900; letter-spacing:-1px; color:var(--text-main); line-height:1; }
		.today-sub{ font-size:0.8rem; color:var(--text-muted); font-weight:600; }
	}

	.heatmap-block { display:flex; }

	.manager-grid {
		display: grid;
		grid-template-columns: 1.15fr 1fr;
		gap: 22px;
		align-items: stretch;
		flex: 1;
		min-height: 0;
	}

	.card {
		background: transparent;
		border: none;
		padding: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		min-height: 0;
	}

	@media (max-width: 980px) {
		.manager-grid { grid-template-columns: 1fr; }
	}
</style>
