<script lang="ts">
	import { getDominantColor } from "@lib/dominantColor";
	import { GetImageBase64, GetExeIcon } from "@lib/api";

	export let recent: any[] = [];
	export let mostPlayed: any[] = [];
	export let stale: any[] = [];
	export let now: number = Date.now();
	export let trackingEnabled: boolean = true;

	// "tabs" = one group at a time, "scroll" = 3 stacked scrollable sections
	let mode: "tabs" | "scroll" = "tabs";
	let tab: "recent" | "played" | "stale" = "recent";

	function formatDuration(sec:number){
		const s=Math.max(0,Math.floor(sec));
		const h=Math.floor(s/3600), m=Math.floor((s%3600)/60);
		if(h>0) return `${h}h ${m}m`;
		if(m>0) return `${m}m`;
		return `${s}s`;
	}
	function formatDate(ts:number){
		return new Intl.DateTimeFormat(undefined,{month:"short",day:"numeric",hour:"2-digit",minute:"2-digit"}).format(new Date(ts));
	}
	function totalSeconds(a:any){
		const active = a.activeSince ? Math.max(0, Math.floor((now - a.activeSince)/1000)) : 0;
		return a.totalPlaytimeSeconds + active;
	}
	function subtitleFor(a:any, which:string){
		if(which==="stale"){
			return a.lastPlayedAt ? `Last ${formatDate(a.lastPlayedAt)} · ${a.sessionCount} sessions` : "Never played";
		}
		return `${formatDate(a.lastPlayedAt)} · ${a.sessionCount} ${a.sessionCount===1?'session':'sessions'}`;
	}

	// icon cache per gamePath
	let iconMap: Record<string,string> = {};
	let colorMap: Record<string,string> = {};

	$: all = [...recent, ...mostPlayed, ...stale];
	$: {
		for(const a of all){
			const key=a.gamePath;
			if(iconMap[key]!==undefined) continue;
			iconMap[key]=""; // mark loading
			loadIcon(a);
		}
	}
	async function loadIcon(a:any){
		try{
			let b64="";
			if(a.customIconPath){ try{ b64=await GetImageBase64(a.customIconPath);}catch{}}
			if(!b64){ try{ b64=await GetExeIcon(a.gamePath);}catch{}}
			if(!b64) return;
			iconMap[a.gamePath]=b64; iconMap={...iconMap};
			const rgb=await getDominantColor(b64);
			if(rgb){ colorMap[a.gamePath]=`${rgb[0]},${rgb[1]},${rgb[2]}`; colorMap={...colorMap}; }
		}catch{}
	}

	$: list = tab==="recent" ? recent : tab==="played" ? mostPlayed : stale;
	$: title = tab==="recent" ? "Recent" : tab==="played" ? "Interest" : "Stale";
	$: sub = tab==="recent"
		? (trackingEnabled ? "Last 7 days · sorted by last played" : "Tracking disabled")
		: tab==="played" ? "By total playtime"
		: "Not played in 30+ days or never launched";

	const sections = [
		{ key:"recent", label:"Recent" },
		{ key:"played", label:"Interest" },
		{ key:"stale", label:"Stale" }
	] as const;
	$: sectionLists = { recent, played: mostPlayed, stale } as Record<string, any[]>;
</script>

<section class="panel">
	<div class="panel-head">
		<div>
			{#if mode==="tabs"}
				<h2>{title}</h2>
				<p>{sub}</p>
			{:else}
				<h2>Library Activity</h2>
				<p>{trackingEnabled ? "Recent · Interest · Stale" : "Tracking disabled"}</p>
			{/if}
		</div>
		<div class="head-controls">
			{#if mode==="tabs"}
				<div class="tabs">
					<button class:active={tab==="recent"} on:click={()=>tab="recent"}>Recent</button>
					<button class:active={tab==="played"} on:click={()=>tab="played"}>Interest</button>
					<button class:active={tab==="stale"} on:click={()=>tab="stale"}>Stale <span class="badge">{stale.length}</span></button>
				</div>
			{/if}
			<div class="mode-switch" role="group" aria-label="View mode">
				<button class:active={mode==="tabs"} on:click={()=>mode="tabs"} title="Tabs view">
					<span class="material-icons">tab</span>
				</button>
				<button class:active={mode==="scroll"} on:click={()=>mode="scroll"} title="Scroll view">
					<span class="material-icons">view_headline</span>
				</button>
			</div>
		</div>
	</div>

	{#if mode==="tabs"}
		{#if list.length>0}
			<div class="list">
				{#each list as a (a.gamePath)}
					{@const rgb = colorMap[a.gamePath] || "255,255,255"}
					{@const icon = iconMap[a.gamePath]}
					<div class="row" style="--row-rgb:{rgb}">
						<div class="row-icon">
							{#if icon}
								<img src={icon} alt="" />
							{:else}
								<span class="material-icons">sports_esports</span>
							{/if}
						</div>
						<div class="row-main">
							<strong>{a.gameName}</strong>
							<span>{subtitleFor(a, tab)}</span>
						</div>
						<div class="row-meta">
							<span class="material-icons">schedule</span>
							{formatDuration(totalSeconds(a))}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty">
				<span class="material-icons">{tab==="stale" ? "history_toggle_off" : "history"}</span>
				<div><strong>No {title.toLowerCase()} games</strong><p>{tab==="stale" ? "All library games were played recently." : "Play some games to build history."}</p></div>
			</div>
		{/if}
	{:else}
		<div class="scroll-sections">
			{#each sections as s}
				{@const items = sectionLists[s.key]}
				<div class="scroll-section">
					<div class="section-title">
						<span>{s.label}</span>
						<span class="count">{items.length}</span>
					</div>
					{#if items.length>0}
						<div class="section-list">
							{#each items as a (a.gamePath)}
								{@const rgb = colorMap[a.gamePath] || "255,255,255"}
								{@const icon = iconMap[a.gamePath]}
								<div class="row compact" style="--row-rgb:{rgb}">
									<div class="row-icon">
										{#if icon}
											<img src={icon} alt="" />
										{:else}
											<span class="material-icons">sports_esports</span>
										{/if}
									</div>
									<div class="row-main">
										<strong>{a.gameName}</strong>
										<span>{subtitleFor(a, s.key)}</span>
									</div>
									<div class="row-meta">
										<span class="material-icons">schedule</span>
										{formatDuration(totalSeconds(a))}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="section-empty">Nothing here yet</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</section>

<style lang="scss">
	.panel{ display:flex; flex-direction:column; gap:16px; flex:1; min-height:0; }
	.panel-head{
		display:flex; align-items:flex-start; justify-content:space-between; gap:12px; flex-wrap:wrap; flex-shrink:0;
		h2{ margin:0; font-size:0.95rem; text-transform:uppercase; letter-spacing:0.8px; }
		p{ margin:4px 0 0; color:var(--text-muted); font-size:0.8rem; }
	}
	.head-controls{ display:flex; align-items:center; gap:10px; flex-wrap:wrap; }
	.tabs{
		display:flex; gap:6px; background:var(--bg-elevated); border:1px solid var(--glass-border); border-radius:var(--radius-pill); padding:4px;
		button{
			border:0; background:transparent; color:var(--text-muted); font-weight:800; font-size:0.75rem;
			padding:6px 12px; border-radius:var(--radius-pill); cursor:pointer; display:flex; align-items:center; gap:6px;
			&.active{ background:var(--accent-primary); color:var(--bg-base); }
			.badge{ background:rgba(255,255,255,0.12); border-radius:999px; padding:1px 6px; font-size:0.68rem; }
			&.active .badge{ background:rgba(0,0,0,0.12); color:inherit; }
		}
	}
	.mode-switch{
		display:flex; gap:2px; background:var(--bg-elevated); border:1px solid var(--glass-border); border-radius:var(--radius-md); padding:3px;
		button{
			border:0; background:transparent; color:var(--text-dim); cursor:pointer;
			padding:4px 8px; border-radius:var(--radius-sm); display:grid; place-items:center;
			.material-icons{ font-size:16px; }
			&.active{ background:var(--accent-primary); color:var(--bg-base); }
		}
	}

	.list{ display:flex; flex-direction:column; gap:8px; flex:1; min-height:0; overflow-y:auto; padding-right:4px; }
	.row{
		display:flex; align-items:center; gap:12px; padding:12px 14px;
		background: var(--bg-elevated); border:1px solid rgba(var(--row-rgb),0.0);
		border-radius:var(--radius-md);
		transition: border-color var(--transition-fast), transform var(--transition-fast);
		&:hover{ border-color: rgba(var(--row-rgb),0.22); transform: translateY(-1px); }
		&.compact{ padding:8px 12px; gap:10px; }
	}
	.row-icon{
		width:40px; height:40px; border-radius:10px; overflow:hidden; background:var(--bg-surface); border:1px solid rgba(255,255,255,0.06);
		display:grid; place-items:center; flex-shrink:0;
		.compact &{ width:32px; height:32px; border-radius:8px; }
		img{ width:100%; height:100%; object-fit:cover; }
		.material-icons{ font-size:18px; color:var(--text-dim); }
	}
	.row-main{ flex:1; min-width:0; display:flex; flex-direction:column; gap:2px; strong{ font-size:0.88rem; color:var(--text-main); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; } span{ font-size:0.74rem; color:var(--text-muted); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; } }
	.row-meta{ display:flex; align-items:center; gap:6px; font-weight:800; font-size:0.8rem; color:var(--text-muted); white-space:nowrap; .material-icons{ font-size:16px; } }

	.scroll-sections{ display:flex; flex-direction:column; gap:16px; flex:1; min-height:0; overflow-y:auto; padding-right:4px; }
	.scroll-section{
		display:flex; flex-direction:column; gap:8px;
		.section-title{
			display:flex; align-items:center; justify-content:space-between;
			font-size:0.72rem; text-transform:uppercase; letter-spacing:1px; font-weight:800; color:var(--text-dim);
			.count{ background:var(--bg-elevated); border:1px solid var(--glass-border); border-radius:999px; padding:1px 8px; font-size:0.66rem; color:var(--text-muted); }
		}
		.section-list{
			display:flex; flex-direction:column; gap:6px;
			max-height: 220px; overflow-y: auto;
			padding-right: 4px;
		}
		.section-empty{
			font-size:0.75rem; color:var(--text-dim); border:1px dashed var(--glass-border); border-radius:var(--radius-md);
			padding:14px; text-align:center;
		}
	}

	.empty{
		display:flex; align-items:center; justify-content:center; gap:12px; min-height:120px;
		border:1px dashed var(--glass-border); border-radius:var(--radius-md); color:var(--text-muted); padding:16px;
		.material-icons{ font-size:26px; color:var(--text-dim); } strong{ color:var(--text-main);} p{ margin:4px 0 0; font-size:0.8rem; }
	}
</style>
