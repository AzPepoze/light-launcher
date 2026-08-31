<script lang="ts">
	export let daySessions: Record<string, { gameName: string; seconds: number }[]> = {};
	export let now: number = Date.now();

	type DaySession = { gameName: string; seconds: number };

	let containerEl: HTMLDivElement;
	let tip: {
		x: number;
		y: number;
		dateText: string;
		total: number;
		games: DaySession[];
	} | null = null;

	function showTip(e: MouseEvent, day: { date: Date; total: number; sessions: DaySession[] }) {
		const target = e.currentTarget as HTMLElement;
		if (!containerEl) return;
		const cRect = containerEl.getBoundingClientRect();
		const tRect = target.getBoundingClientRect();
		tip = {
			x: tRect.left - cRect.left + tRect.width / 2,
			y: tRect.top - cRect.top,
			dateText: day.date.toLocaleDateString(undefined, { month: "short", day: "numeric" }),
			total: day.total,
			games: day.sessions
		};
	}

	function hideTip() {
		tip = null;
	}

	function fmt(sec: number): string {
		const s = Math.max(0, Math.floor(sec));
		const h = Math.floor(s / 3600);
		const m = Math.floor((s % 3600) / 60);
		if (h > 0) return `${h}h ${m}m`;
		if (m > 0) return `${m}m`;
		return `${s}s`;
	}

	function dayKeyOf(ts: number): string {
		const d = new Date(ts);
		d.setHours(0, 0, 0, 0);
		return d.toISOString().slice(0, 10);
	}

	// level from total playtime that day: 0 none · 1 <30m · 2 <1h · 3 <2h · 4 >=2h
	$: heatmapDays = (() => {
		const days: { date: Date; level: number; total: number; sessions: DaySession[] }[] = [];
		const today = new Date(now);
		today.setHours(0,0,0,0);
		const start = new Date(today);
		start.setDate(today.getDate() - 181); // 182 days = 26 weeks
		for (let i=0;i<182;i++) {
			const d = new Date(start);
			d.setDate(start.getDate() + i);
			const sessions = daySessions[dayKeyOf(d.getTime())] || [];
			const total = sessions.reduce((s, g) => s + g.seconds, 0);
			let level = 0;
			if (total >= 2 * 3600) level = 4;
			else if (total >= 3600) level = 3;
			else if (total >= 1800) level = 2;
			else if (total > 0) level = 1;
			days.push({ date: d, level, total, sessions });
		}
		return days;
	})();

	$: weeks = (() => {
		const w: typeof heatmapDays[] = [];
		for (let i=0;i<heatmapDays.length;i+=7) w.push(heatmapDays.slice(i,i+7));
		return w;
	})();
</script>

<div class="heatmap" bind:this={containerEl} on:mouseleave={hideTip}>
	<div class="heatmap-grid">
		{#each weeks as week}
			<div class="week-col">
				{#each week as day}
					<div
						class="day-cell level-{day.level}"
						on:mouseenter={(e) => showTip(e, day)}
						on:mouseleave={hideTip}
					></div>
				{/each}
			</div>
		{/each}
	</div>
	<div class="heatmap-legend">
		<span>Less</span>
		<div class="legend-cells">
			<div class="day-cell level-0 sm"></div>
			<div class="day-cell level-1 sm"></div>
			<div class="day-cell level-2 sm"></div>
			<div class="day-cell level-3 sm"></div>
			<div class="day-cell level-4 sm"></div>
		</div>
		<span>More</span>
	</div>

	{#if tip}
		<div class="heatmap-tip" style="left:{tip.x}px; top:{tip.y}px;">
			<div class="tip-header">
				<span class="tip-date">{tip.dateText}</span>
				<span class="tip-total">{tip.total > 0 ? fmt(tip.total) : "No playtime"}</span>
			</div>
			{#if tip.games.length > 0}
				<div class="tip-games">
					{#each tip.games.slice(0, 6) as g}
						<div class="tip-row">
							<span class="tip-name">{g.gameName}</span>
							<span class="tip-time">{fmt(g.seconds)}</span>
						</div>
					{/each}
					{#if tip.games.length > 6}
						<div class="tip-more">+{tip.games.length - 6} more</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.heatmap {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-end;
	}
	.heatmap-grid { display:flex; gap:3px; }
	.week-col { display:flex; flex-direction:column; gap:3px; }
	.day-cell {
		width: 11px; height: 11px; border-radius:3px;
		background: rgba(255,255,255,0.05);
		transition: transform 100ms ease;
		cursor: default;
		&:hover { transform: scale(1.35); }
		/* GitHub contribution colors (dark) */
		&.level-0{ background: rgba(255,255,255,0.05); }
		&.level-1{ background:#0e4429; }
		&.level-2{ background:#006d32; }
		&.level-3{ background:#26a641; }
		&.level-4{ background:#39d353; }
		&.sm{ width:10px; height:10px; }
	}
	.heatmap-legend{
		display:flex; align-items:center; gap:6px; font-size:0.68rem; color:var(--text-dim); font-weight:700;
		.legend-cells{ display:flex; gap:3px; }
	}

	.heatmap-tip {
		position: absolute;
		transform: translate(-50%, calc(-100% - 8px));
		background: #1c2128;
		border: 1px solid rgba(255,255,255,0.12);
		color: #e6edf3;
		padding: 8px 12px;
		border-radius: 8px;
		white-space: nowrap;
		pointer-events: none;
		z-index: 20;
		box-shadow: 0 4px 14px rgba(0,0,0,0.4);
		animation: tipIn 120ms ease both;

		&::after {
			content: "";
			position: absolute;
			top: 100%;
			left: 50%;
			transform: translateX(-50%);
			border: 5px solid transparent;
			border-top-color: #1c2128;
		}
	}

	.tip-header {
		display: flex;
		align-items: baseline;
		gap: 10px;
		padding-bottom: 4px;
		margin-bottom: 4px;
		border-bottom: 1px solid rgba(255,255,255,0.08);

		.tip-date { font-size: 0.72rem; font-weight: 800; }
		.tip-total { font-size: 0.72rem; font-weight: 800; color: #3fb950; }
	}

	.tip-games {
		display: flex;
		flex-direction: column;
		gap: 2px;

		.tip-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 16px;
			font-size: 0.7rem;

			.tip-name { color: #b6c2cf; overflow: hidden; text-overflow: ellipsis; max-width: 180px; }
			.tip-time { font-weight: 800; color: #e6edf3; }
		}

		.tip-more {
			font-size: 0.66rem;
			color: #7d8590;
			margin-top: 2px;
		}
	}

	@keyframes tipIn {
		from { opacity: 0; transform: translate(-50%, calc(-100% - 4px)); }
		to { opacity: 1; transform: translate(-50%, calc(-100% - 8px)); }
	}
</style>
