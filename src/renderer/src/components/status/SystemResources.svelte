<script lang="ts">
	import MiniResourceGraph from "./MiniResourceGraph.svelte";
	import type { SystemInfo, SystemUsage } from "@shared";

	interface Props {
		sysInfo: SystemInfo;
		sysUsage: SystemUsage;
		cpuHistory: number[];
		ramHistory: number[];
		gpuHistoryMap: Record<number, number[]>;
	}

	let { sysInfo, sysUsage, cpuHistory, ramHistory, gpuHistoryMap }: Props = $props();

	let gpuIdx = $state(0);
	let slideDir = $state(1);

	const gpuList = $derived(
		sysInfo.gpus?.length ? sysInfo.gpus : sysInfo.gpu ? [sysInfo.gpu] : []
	);
	const safeGpuIdx = $derived(Math.max(0, Math.min(gpuIdx, gpuList.length - 1)));
	const currentGpu = $derived(gpuList[safeGpuIdx] ?? "");
	const currentGpuUsage = $derived(sysUsage.gpus[safeGpuIdx] ?? 0);
	const currentGpuHistory = $derived(gpuHistoryMap[safeGpuIdx] ?? []);

	function nextGpu() {
		if (gpuList.length < 2) return;
		slideDir = 1;
		gpuIdx = (safeGpuIdx + 1) % gpuList.length;
	}

	function prevGpu() {
		if (gpuList.length < 2) return;
		slideDir = -1;
		gpuIdx = (safeGpuIdx - 1 + gpuList.length) % gpuList.length;
	}
</script>

<div class="status-grid">
	<!-- CPU -->
	<div class="status-box">
		<div class="box-header">
			<div class="header-left">
				<span class="material-icons mini-icon">memory</span>
				<span class="label">CPU</span>
			</div>
			<span class="usage">{sysUsage.cpu.toFixed(1)}%</span>
		</div>
		<MiniResourceGraph data={cpuHistory} color="var(--accent-primary, #64ffda)" height={46} />
		<span class="info-text" title={sysInfo.cpu}>{sysInfo.cpu}</span>
	</div>

	<!-- RAM -->
	<div class="status-box">
		<div class="box-header">
			<div class="header-left">
				<span class="material-icons mini-icon">storage</span>
				<span class="label">RAM</span>
			</div>
			<span class="usage">{sysUsage.ram}%</span>
		</div>
		<MiniResourceGraph data={ramHistory} color="var(--accent-secondary, #b197fc)" height={46} />
		<span class="info-text">{sysUsage.ramUsedGb} GB used</span>
	</div>

	<!-- GPU -->
	<div class="status-box">
		<div class="box-header">
			<div class="header-left">
				<span class="material-icons mini-icon">videogame_asset</span>
				<span class="label">GPU</span>
			</div>
			<span class="usage">{currentGpuUsage}%</span>
		</div>
		<MiniResourceGraph data={currentGpuHistory} color="#10b981" height={46} />

		<div class="gpu-slider">
			{#if gpuList.length > 1}
				<button class="gpu-arrow left" onclick={prevGpu} aria-label="Previous GPU">
					<span class="material-icons">chevron_left</span>
				</button>
			{/if}
			<div class="gpu-name-wrap">
				{#key safeGpuIdx}
					<span
						class="info-text gpu-name"
						title="{currentGpu} ({sysInfo.driver})"
						style="animation: {slideDir > 0 ? 'slideInRight' : 'slideInLeft'} 0.22s cubic-bezier(0.16, 1, 0.3, 1)"
					>
						{currentGpu}
					</span>
				{/key}
			</div>
			{#if gpuList.length > 1}
				<button class="gpu-arrow right" onclick={nextGpu} aria-label="Next GPU">
					<span class="material-icons">chevron_right</span>
				</button>
			{/if}
		</div>

		{#if gpuList.length > 1}
			<div class="gpu-dots">
				{#each gpuList as _, i}
					<span class="gpu-dot" class:active={i === safeGpuIdx}></span>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.status-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 20px;
	}

	.status-box {
		background: var(--bg-base);
		padding: 16px;
		border-radius: var(--radius-md);
		border: 1px solid rgba(255, 255, 255, 0.03);
		display: flex;
		flex-direction: column;
		gap: 8px;
		transition: border-color var(--transition-fast), background var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			background: rgba(255, 255, 255, 0.02);
		}

		.box-header {
			display: flex;
			align-items: center;
			justify-content: space-between;

			.header-left {
				display: flex;
				align-items: center;
				gap: 8px;
			}

			.mini-icon {
				font-size: 16px;
				color: var(--text-muted);
			}

			.label {
				font-size: 0.8rem;
				font-weight: 700;
				color: var(--text-muted);
				text-transform: uppercase;
				letter-spacing: 0.5px;
			}

			.usage {
				font-size: 0.9rem;
				font-weight: 800;
				font-family: monospace;
				color: var(--accent-primary);
			}
		}

		.info-text {
			font-size: 0.78rem;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			display: block;
			margin-top: 2px;
		}

		.gpu-slider {
			display: flex;
			align-items: center;
			gap: 4px;
			position: relative;
			min-width: 0;

			.gpu-arrow {
				background: transparent;
				border: none;
				color: var(--text-muted);
				cursor: pointer;
				padding: 0;
				display: flex;
				align-items: center;
				justify-content: center;
				flex-shrink: 0;
				border-radius: 4px;
				transition: color 0.15s, background 0.15s;

				&:hover {
					color: var(--text-main);
					background: rgba(255, 255, 255, 0.08);
				}

				.material-icons {
					font-size: 16px;
				}
			}

			.gpu-name-wrap {
				flex: 1;
				min-width: 0;
				overflow: hidden;
			}

			.gpu-name {
				display: block;
				width: 100%;
			}
		}

		.gpu-dots {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 4px;
			margin-top: 2px;

			.gpu-dot {
				width: 5px;
				height: 5px;
				border-radius: 50%;
				background: rgba(255, 255, 255, 0.2);
				transition: background var(--transition-fast), width var(--transition-fast);

				&.active {
					background: #10b981;
					width: 14px;
				}
			}
		}
	}

	@keyframes slideInRight {
		from {
			opacity: 0;
			transform: translateX(12px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes slideInLeft {
		from {
			opacity: 0;
			transform: translateX(-12px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}
</style>
