<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import StatusDrawer from "@components/shared/StatusDrawer.svelte";
	import { GetRunningSessions, KillSession, GetAppSettings } from "@lib/api";
	import { GetGameActivity } from "@lib/activityApi";
	import { notifications } from "@stores/notificationStore";
	import type { AppSettings, GameActivity, RunningSession } from "@shared";

	let sessions: RunningSession[] = [];
	let activities: GameActivity[] = [];
	let appSettings: AppSettings | null = null;
	let refreshTimer: ReturnType<typeof setInterval> | null = null;
	let clockTimer: ReturnType<typeof setInterval> | null = null;
	let now = Date.now();

	$: recentActivities = activities.slice(0, 12);
	$: trackingEnabled = appSettings?.TrackPlaytime !== false;

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
			const [running, history, settings] = await Promise.all([
				GetRunningSessions(),
				GetGameActivity(),
				GetAppSettings()
			]);
			sessions = running || [];
			activities = history || [];
			appSettings = settings;
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

	function activityFor(session: RunningSession): GameActivity | undefined {
		return activities.find((item) => item.gamePath === session.gamePath);
	}

	function sessionStartedAt(session: RunningSession): number {
		return activityFor(session)?.activeSince || session.startedAt || now;
	}

	function totalSeconds(activity: GameActivity): number {
		const active = activity.activeSince ? Math.max(0, Math.floor((now - activity.activeSince) / 1000)) : 0;
		return activity.totalPlaytimeSeconds + active;
	}

	function formatDuration(totalSeconds: number): string {
		const seconds = Math.max(0, Math.floor(totalSeconds));
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		if (hours > 0) return `${hours}h ${minutes}m`;
		if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
		return `${seconds}s`;
	}

	function formatDate(timestamp: number): string {
		return new Intl.DateTimeFormat(undefined, {
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit"
		}).format(new Date(timestamp));
	}
</script>

<div class="manager-container">
	<PageHeader title="Game Manager" icon="monitor_heart" />

	<section class="manager-section">
		<div class="section-heading">
			<div>
				<h2>Running Games</h2>
				<p>Live sessions managed by LightLauncher.</p>
			</div>
			<span class="count-badge">{sessions.length}</span>
		</div>

		{#if sessions.length > 0}
			<div class="running-list">
				{#each sessions as session}
					<div class="running-card">
						<div class="running-icon"><span class="material-icons">sports_esports</span></div>
						<div class="running-info">
							<strong>{session.gameName}</strong>
							<span class="path" title={session.gamePath}>{session.gamePath}</span>
							<span class="runtime">Running for {formatDuration((now - sessionStartedAt(session)) / 1000)} · PID {session.pid}</span>
						</div>
						<button class="stop-btn" on:click={() => stopSession(session)} title="Stop game">
							<span class="material-icons">stop</span>
							Stop
						</button>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-state">
				<span class="material-icons">bedtime</span>
				<div><strong>No managed games are running</strong><p>Games launched from LightLauncher will appear here.</p></div>
			</div>
		{/if}
	</section>

	<section class="manager-section">
		<div class="section-heading">
			<div>
				<h2>Recent & Playtime</h2>
				<p>{trackingEnabled ? "Recent sessions and accumulated playtime." : "Playtime tracking is disabled in Settings."}</p>
			</div>
			<span class="material-icons heading-icon">history</span>
		</div>

		{#if recentActivities.length > 0}
			<div class="activity-list">
				{#each recentActivities as activity}
					<div class="activity-row">
						<div class="activity-main">
							<strong>{activity.gameName}</strong>
							<span>{formatDate(activity.lastPlayedAt)} · {activity.sessionCount} {activity.sessionCount === 1 ? 'session' : 'sessions'}</span>
						</div>
						<div class="playtime">
							<span class="material-icons">schedule</span>
							{formatDuration(totalSeconds(activity))}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-state compact">
				<span class="material-icons">history_toggle_off</span>
				<div><strong>No play history yet</strong><p>Tracked launches will show up here.</p></div>
			</div>
		{/if}
	</section>

	<section class="manager-section status-section">
		<div class="section-heading">
			<div>
				<h2>System Status & Utilities</h2>
				<p>Resource monitoring and maintenance tools, now kept with game management.</p>
			</div>
			<span class="material-icons heading-icon">tune</span>
		</div>
		<StatusDrawer embedded={true} />
	</section>
</div>

<style lang="scss">
	.manager-container {
		display: flex;
		flex-direction: column;
		gap: 28px;
		padding-bottom: 32px;
	}

	.manager-section {
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		padding: 28px;
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.18);
	}

	.section-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		margin-bottom: 20px;

		h2 { margin: 0; font-size: 1.1rem; text-transform: uppercase; letter-spacing: 0.8px; }
		p { margin: 5px 0 0; color: var(--text-muted); font-size: 0.88rem; }
	}

	.count-badge, .heading-icon {
		color: var(--accent-primary);
		font-weight: 800;
	}
	.count-badge { background: var(--bg-elevated); border: 1px solid var(--glass-border); border-radius: var(--radius-pill); padding: 6px 12px; }

	.running-list, .activity-list { display: flex; flex-direction: column; gap: 10px; }

	.running-card, .activity-row {
		display: flex;
		align-items: center;
		gap: 14px;
		background: var(--bg-elevated);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 14px 16px;
	}

	.running-icon {
		width: 42px; height: 42px; border-radius: var(--radius-md); display: grid; place-items: center;
		background: var(--accent-glow); color: var(--accent-primary); flex-shrink: 0;
	}

	.running-info, .activity-main { display: flex; flex-direction: column; gap: 4px; min-width: 0; flex: 1; }
	.running-info strong, .activity-main strong { color: var(--text-main); }
	.path { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-dim); font-family: monospace; font-size: 0.75rem; }
	.runtime, .activity-main span { color: var(--text-muted); font-size: 0.8rem; }

	.stop-btn {
		display: inline-flex; align-items: center; gap: 6px; border: 1px solid rgba(255, 80, 80, 0.25);
		background: rgba(255, 80, 80, 0.08); color: var(--danger, #ff5f5f); border-radius: var(--radius-md);
		padding: 9px 12px; cursor: pointer; font-weight: 800;
		&:hover { background: rgba(255, 80, 80, 0.16); }
		.material-icons { font-size: 18px; }
	}

	.playtime { display: flex; align-items: center; gap: 7px; color: var(--accent-primary); font-weight: 800; white-space: nowrap; }
	.playtime .material-icons { font-size: 18px; }

	.empty-state {
		display: flex; align-items: center; justify-content: center; gap: 14px; min-height: 130px;
		border: 1px dashed var(--glass-border); border-radius: var(--radius-md); color: var(--text-muted);
		.material-icons { font-size: 30px; color: var(--text-dim); }
		strong { color: var(--text-main); }
		p { margin: 4px 0 0; font-size: 0.82rem; }
		&.compact { min-height: 100px; }
	}

	.status-section { padding-bottom: 28px; }
</style>
