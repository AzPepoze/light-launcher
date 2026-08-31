<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import StatusDrawer from "@components/status/StatusDrawer.svelte";
	import HeatmapCard from "@components/manager/HeatmapCard.svelte";
	import RunningRail from "@components/manager/RunningRail.svelte";
	import ActivityTabs from "@components/manager/ActivityTabs.svelte";
	import { GetRunningSessions, KillSession, GetAppSettings, GetGameActivity, GetAllGames, GetAutoScannedGames } from "@lib/api";
	import { notifications } from "@stores/notificationStore";
	import type { AppSettings, GameActivity, RunningSession, GameInfo, ScannedFolderGroup } from "@shared";

	const DAY = 24 * 60 * 60 * 1000;

	type DaySession = { gameName: string; seconds: number };
	type DaySessionMap = Record<string, DaySession[]>;

	function dayKeyOf(ts: number): string {
		const d = new Date(ts);
		d.setHours(0, 0, 0, 0);
		return d.toISOString().slice(0, 10);
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
				out.push({ gamePath:g.path, gameName:g.name, lastPlayedAt:0, totalPlaytimeSeconds:0 } as GameActivity);
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
