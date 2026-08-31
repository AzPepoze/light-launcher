<script lang="ts">
	import { getDominantColor } from "@lib/dominantColor";
	import { GetImageBase64, GetExeIcon, GetConfig } from "@lib/api";
	import type { GameActivity, RunningSession } from "@shared";
	import type { LaunchOptions } from "@shared";

	export let sessions: RunningSession[] = [];
	export let activities: GameActivity[] = [];
	export let now: number = Date.now();
	export let onStop: (runningSession: RunningSession) => void = () => {};

	type RuntimeDetails = {
		prefixName: string;
		protonName: string;
		prefixPath: string;
		protonPath: string;
	};

	function activityFor(runningSession: RunningSession): GameActivity | undefined {
		return activities.find((activity) => activity.gamePath === runningSession.gamePath);
	}

	function sessionStartedAt(runningSession: RunningSession): number {
		return activityFor(runningSession)?.activeSince || runningSession.startedAt || now;
	}

	function formatDuration(totalSeconds: number): string {
		const clampedSeconds = Math.max(0, Math.floor(totalSeconds));
		const hours = Math.floor(clampedSeconds / 3600);
		const minutes = Math.floor((clampedSeconds % 3600) / 60);
		if (hours > 0) return `${hours}h ${minutes}m`;
		if (minutes > 0) return `${minutes}m ${clampedSeconds % 60}s`;
		return `${clampedSeconds}s`;
	}

	function extractBaseName(filePath: string): string {
		if (!filePath) return "";
		const normalizedPath = filePath.replace(/\\/g, "/");
		const segments = normalizedPath.split("/").filter((segment) => segment.length > 0);
		return segments.length > 0 ? segments[segments.length - 1] : filePath;
	}

	function runtimeDetailsFromLaunchOptions(launchOptions: LaunchOptions | null): RuntimeDetails {
		if (!launchOptions) {
			return { prefixName: "Default", protonName: "—", prefixPath: "", protonPath: "" };
		}
		const prefixPath = launchOptions.PrefixPath ?? "";
		const protonPath = launchOptions.ProtonPath ?? "";
		const prefixName = prefixPath ? extractBaseName(prefixPath) : "Default";
		const protonName = protonPath ? extractBaseName(protonPath) : "—";
		return { prefixName, protonName, prefixPath, protonPath };
	}

	let icons: Record<number, string> = {};
	let colors: Record<number, string> = {};
	let runtimeDetailsByPid: Record<number, RuntimeDetails> = {};

	function runtimeDetailsForSession(runningSession: RunningSession): RuntimeDetails {
		const cachedDetails = runtimeDetailsByPid[runningSession.pid];
		if (cachedDetails) return cachedDetails;

		const extendedSession = runningSession as RunningSession & {
			prefixPath?: string;
			protonPath?: string;
			protonName?: string;
		};
		const hasEmbeddedRuntime = Boolean(extendedSession.prefixPath || extendedSession.protonPath || extendedSession.protonName);
		if (hasEmbeddedRuntime) {
			const prefixPath = extendedSession.prefixPath ?? "";
			const protonPath = extendedSession.protonPath ?? "";
			const prefixName = prefixPath ? extractBaseName(prefixPath) : "Default";
			const protonName = extendedSession.protonName ?? (protonPath ? extractBaseName(protonPath) : "—");
			return { prefixName, protonName, prefixPath, protonPath };
		}

		return { prefixName: "Default", protonName: "—", prefixPath: "", protonPath: "" };
	}

	async function loadIcon(runningSession: RunningSession): Promise<void> {
		try {
			const activity = activityFor(runningSession);
			let iconBase64 = "";
			if (activity?.customIconPath) {
				try {
					iconBase64 = await GetImageBase64(activity.customIconPath);
				} catch {}
			}
			if (!iconBase64) {
				try {
					iconBase64 = await GetExeIcon(runningSession.gamePath);
				} catch {}
			}
			if (!iconBase64) return;
			icons[runningSession.pid] = iconBase64;
			icons = { ...icons };
			const dominantRgb = await getDominantColor(iconBase64);
			if (dominantRgb) {
				colors[runningSession.pid] = `${dominantRgb[0]},${dominantRgb[1]},${dominantRgb[2]}`;
				colors = { ...colors };
			}
		} catch {}
	}

	async function loadRuntimeDetails(runningSession: RunningSession): Promise<void> {
		if (runtimeDetailsByPid[runningSession.pid]) return;

		const extendedSession = runningSession as RunningSession & {
			prefixPath?: string;
			protonPath?: string;
			protonName?: string;
		};
		if (extendedSession.prefixPath || extendedSession.protonPath || extendedSession.protonName) {
			const prefixPath = extendedSession.prefixPath ?? "";
			const protonPath = extendedSession.protonPath ?? "";
			const prefixName = prefixPath ? extractBaseName(prefixPath) : "Default";
			const protonName = extendedSession.protonName ?? (protonPath ? extractBaseName(protonPath) : "—");
			runtimeDetailsByPid[runningSession.pid] = { prefixName, protonName, prefixPath, protonPath };
			runtimeDetailsByPid = { ...runtimeDetailsByPid };
			return;
		}

		try {
			const launchOptions = await GetConfig(runningSession.gamePath);
			runtimeDetailsByPid[runningSession.pid] = runtimeDetailsFromLaunchOptions(launchOptions);
			runtimeDetailsByPid = { ...runtimeDetailsByPid };
		} catch {}
	}

	$: {
		for (const runningSession of sessions) {
			if (!icons[runningSession.pid]) {
				void loadIcon(runningSession);
			}
			if (!runtimeDetailsByPid[runningSession.pid]) {
				void loadRuntimeDetails(runningSession);
			}
		}
	}

	$: {
		const activePids = new Set(sessions.map((runningSession) => runningSession.pid));
		let hasStaleEntry = false;
		for (const pidKey of Object.keys(runtimeDetailsByPid)) {
			const numericPid = Number(pidKey);
			if (!activePids.has(numericPid)) {
				delete runtimeDetailsByPid[numericPid];
				hasStaleEntry = true;
			}
		}
		for (const pidKey of Object.keys(icons)) {
			const numericPid = Number(pidKey);
			if (!activePids.has(numericPid)) {
				delete icons[numericPid];
				delete colors[numericPid];
				hasStaleEntry = true;
			}
		}
		if (hasStaleEntry) {
			runtimeDetailsByPid = { ...runtimeDetailsByPid };
			icons = { ...icons };
			colors = { ...colors };
		}
	}
</script>

<section class="rail">
	<div class="rail-header">
		<div>
			<h2>Running</h2>
			<p>Live sessions with gradient tint</p>
		</div>
		<span class="count">{sessions.length}</span>
	</div>

	{#if sessions.length > 0}
		<div class="rail-list">
			{#each sessions as runningSession (runningSession.pid)}
				{@const rgb = colors[runningSession.pid] || "255,255,255"}
				{@const runtimeDetails = runtimeDetailsByPid[runningSession.pid] ?? runtimeDetailsForSession(runningSession)}
				<div class="running-card" style="--card-rgb: {rgb}">
					<div class="card-bg"></div>
					<div class="card-inner">
						<div class="icon-wrap">
							<div class="icon-tilt">
								{#if icons[runningSession.pid]}
									<img src={icons[runningSession.pid]} alt="" />
								{:else}
									<span class="material-icons fallback">sports_esports</span>
								{/if}
							</div>
						</div>
						<div class="info">
							<strong class="name">{runningSession.gameName}</strong>
							<span class="path" title={runningSession.gamePath}>{runningSession.gamePath}</span>
							<span
								class="meta"
								title={[runtimeDetails.prefixPath ? `Prefix: ${runtimeDetails.prefixPath}` : "", runtimeDetails.protonPath ? `Proton: ${runtimeDetails.protonPath}` : ""]
									.filter(Boolean)
									.join(" · ") || runtimeDetails.prefixName}
							>
								<span>Prefix: {runtimeDetails.prefixName}</span>
								<span class="dot">·</span>
								<span>Proton: {runtimeDetails.protonName}</span>
							</span>
							<span class="runtime">Running for {formatDuration((now - sessionStartedAt(runningSession)) / 1000)} · PID {runningSession.pid}</span>
						</div>
						<button
							class="stop-btn"
							on:click={() => onStop(runningSession)}
							title="Stop {runningSession.gameName}"
							aria-label={`Stop ${runningSession.gameName}`}
						>
							<span class="material-icons">close</span>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="empty">
			<span class="material-icons">bedtime</span>
			<div><strong>No managed games are running</strong><p>Launch from LightLauncher to appear here.</p></div>
		</div>
	{/if}
</section>

<style lang="scss">
	.rail { display:flex; flex-direction:column; gap:16px; flex:1; min-height:0; }
	.rail-header {
		display:flex; align-items:center; justify-content:space-between; gap:12px; flex-shrink:0;
		h2{ margin:0; font-size:0.95rem; text-transform:uppercase; letter-spacing:0.8px; }
		p{ margin:4px 0 0; color:var(--text-muted); font-size:0.8rem; }
		.count{ background:var(--bg-elevated); border:1px solid var(--glass-border); border-radius:var(--radius-pill); padding:6px 12px; font-weight:800; color:var(--accent-primary); }
	}

	.rail-list{ display:flex; flex-direction:column; gap:12px; flex:1; min-height:0; overflow-y:auto; padding-right:4px; }

	.running-card{
		position:relative; border-radius:var(--radius-lg); overflow:hidden;
		border:1px solid rgba(var(--card-rgb),0.22);
		background: var(--bg-surface);
		box-shadow: 0 10px 30px rgba(0,0,0,0.22);
		transition: transform var(--transition-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);
		&:hover{ transform: translateY(-2px); border-color: rgba(var(--card-rgb),0.38); box-shadow: 0 14px 36px rgba(0,0,0,0.28); }
	}

	.card-bg{
		position:absolute; inset:0;
		background: linear-gradient(135deg, rgba(var(--card-rgb),0.22) 0%, rgba(var(--card-rgb),0.06) 42%, transparent 72%);
		pointer-events:none;
	}

	.card-inner{
		position:relative; display:flex; align-items:center; gap:20px; padding:20px;
	}

	.icon-wrap{
		width:160px; height:160px; flex-shrink:0; perspective:900px; display:grid; place-items:center;
	}

	.icon-tilt{
		width:160px; height:160px; border-radius:24px; overflow:hidden; background: var(--bg-elevated);
		border:1px solid rgba(var(--card-rgb),0.25);
		transform: rotateY(14deg) rotateX(4deg);
		box-shadow: 0 16px 36px rgba(0,0,0,0.38);
		display:grid; place-items:center;
		transition: transform var(--transition-spring);
		img{ width:100%; height:100%; object-fit:cover; }
		.fallback{ color:var(--text-dim); font-size:64px; }
		.running-card:hover &{ transform: rotateY(4deg) rotateX(0deg) scale(1.02); }
	}

	.info{ flex:1; min-width:0; display:flex; flex-direction:column; gap:5px; }
	.name{ color:var(--text-main); font-size:1.5rem; font-weight:900; letter-spacing:-0.4px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
	.path{ color:var(--text-dim); font-family:monospace; font-size:0.72rem; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
	.meta{
		color:var(--text-muted);
		font-size:0.78rem;
		font-weight:600;
		display:flex;
		align-items:center;
		gap:6px;
		flex-wrap:wrap;
		.dot{ color:var(--text-dim); }
	}
	.runtime{ color:var(--text-muted); font-size:0.85rem; font-weight:700; }

	.stop-btn{
		display:grid; place-items:center; border:1px solid rgba(255,80,80,0.22);
		background: rgba(255,80,80,0.08); color:var(--danger); border-radius:999px;
		width:34px; height:34px; padding:0; cursor:pointer; flex-shrink:0;
		transition: background var(--transition-fast), transform var(--transition-fast);
		&:hover{ background: rgba(255,80,80,0.16); transform: scale(1.06); }
		&:active{ transform: scale(0.94); }
		.material-icons{ font-size:18px; }
	}

	.empty{
		display:flex; align-items:center; justify-content:center; gap:12px; min-height:120px;
		border:1px dashed var(--glass-border); border-radius:var(--radius-md); color:var(--text-muted); padding:16px;
		.material-icons{ font-size:28px; color:var(--text-dim); }
		strong{ color:var(--text-main); } p{ margin:4px 0 0; font-size:0.8rem; }
	}
</style>
