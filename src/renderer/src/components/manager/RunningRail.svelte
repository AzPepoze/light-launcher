<script lang="ts">
	import { onMount } from "svelte";
	import { getDominantColor } from "@lib/dominantColor";
	import { GetImageBase64, GetExeIcon } from "@lib/api";
	import type { GameActivity, RunningSession } from "@shared";

	export let sessions: RunningSession[] = [];
	export let activities: GameActivity[] = [];
	export let now: number = Date.now();
	export let onStop: (s: RunningSession)=>void = ()=>{};

	function activityFor(session: RunningSession) { return activities.find((activity)=> activity.gamePath === session.gamePath); }
	function sessionStartedAt(session: RunningSession) { return activityFor(session)?.activeSince || session.startedAt || now; }
	function formatDuration(totalSeconds:number){
		const clampedSeconds=Math.max(0,Math.floor(totalSeconds));
		const hours=Math.floor(clampedSeconds/3600), minutes=Math.floor((clampedSeconds%3600)/60);
		if(hours>0) return `${hours}h ${minutes}m`;
		if(minutes>0) return `${minutes}m ${clampedSeconds%60}s`;
		return `${clampedSeconds}s`;
	}

	// icon + color per session
	let icons: Record<number,string> = {};
	let colors: Record<number,string> = {};

	$: {
		// trigger load when sessions change
		for(const session of sessions){
			if(icons[session.pid]) continue;
			loadIcon(session);
		}
	}

	async function loadIcon(session: RunningSession){
		try{
			const activity = activityFor(session);
			let iconBase64 = "";
			if(activity?.customIconPath){
				try{ iconBase64 = await GetImageBase64(activity.customIconPath);}catch{}
			}
			if(!iconBase64){
				try{ iconBase64 = await GetExeIcon(session.gamePath);}catch{}
			}
			if(!iconBase64) return;
			icons[session.pid]=iconBase64;
			icons={...icons};
			const dominantRgb = await getDominantColor(iconBase64);
			if(dominantRgb) {
				colors[session.pid]=`${dominantRgb[0]},${dominantRgb[1]},${dominantRgb[2]}`;
				colors={...colors};
			}
		}catch{}
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

	{#if sessions.length>0}
		<div class="rail-list">
			{#each sessions as session (session.pid)}
				{@const rgb = colors[session.pid] || "255,255,255"}
				<div class="running-card" style="--card-rgb: {rgb}">
					<div class="card-bg"></div>
					<div class="card-inner">
						<div class="icon-wrap">
							<div class="icon-tilt">
								{#if icons[session.pid]}
									<img src={icons[session.pid]} alt="" />
								{:else}
									<span class="material-icons fallback">sports_esports</span>
								{/if}
							</div>
						</div>
						<div class="info">
							<strong class="name">{session.gameName}</strong>
							<span class="path" title={session.gamePath}>{session.gamePath}</span>
							<span class="runtime">Running for {formatDuration((now - sessionStartedAt(session))/1000)} · PID {session.pid}</span>
						</div>
						<button class="stop-btn" on:click={()=>onStop(session)} title="Stop game" aria-label={`Stop ${session.gameName}`}>
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
