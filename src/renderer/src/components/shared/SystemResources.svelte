<script lang="ts">
	import { fly } from "svelte/transition";
	import MiniResourceGraph from "@components/shared/MiniResourceGraph.svelte";
	import { ResourceHistoryTracker } from "@components/shared/ResourceHistoryTracker.svelte";

	export let sysInfo: {
		os: string;
		kernel: string;
		cpu: string;
		gpu: string;
		gpus?: string[];
		ram: string;
		driver: string;
	};
	export let sysUsage: {
		cpu: string;
		ram: string;
		gpu: string;
		gpus?: string[];
	};

	const history = new ResourceHistoryTracker(30);

	$: if (sysUsage) {
		history.push(sysUsage);
	}

	$: gpuList = sysInfo.gpus?.length ? sysInfo.gpus : sysInfo.gpu ? [sysInfo.gpu] : [];

	let gpuIdx = 0;
	let slideDir = 1;

	$: if (gpuIdx >= gpuList.length) gpuIdx = 0;
	$: currentGpu = gpuList[gpuIdx] ?? "";
	$: currentGpuUsage = sysUsage.gpus?.[gpuIdx] ?? sysUsage.gpu ?? "0%";
	$: currentGpuHistory = history.getGpuHistory(gpuIdx);

	function nextGpu() {
		if (gpuList.length < 2) return;
		slideDir = 1;
		gpuIdx = (gpuIdx + 1) % gpuList.length;
	}

	function prevGpu() {
		if (gpuList.length < 2) return;
		slideDir = -1;
		gpuIdx = (gpuIdx - 1 + gpuList.length) % gpuList.length;
	}
</script>

<div class="status-grid">
	<!-- OS -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">laptop_windows</span>
				<span class="label">SYSTEM</span>
			</div>
		</div>
		<div class="system-info">
			<span class="os-text" title={sysInfo.os}>{sysInfo.os}</span>
			<span class="kernel-text">Kernel: {sysInfo.kernel}</span>
		</div>
	</div>

	<!-- CPU -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">memory</span>
				<span class="label">CPU</span>
			</div>
			<span class="usage">{sysUsage.cpu}</span>
		</div>
		<MiniResourceGraph
			data={history.cpu}
			color="var(--accent-primary, #64ffda)"
			height={46}
		/>
		<span class="info-text" title={sysInfo.cpu}>{sysInfo.cpu}</span>
	</div>

	<!-- RAM -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">storage</span>
				<span class="label">RAM</span>
			</div>
			<span class="usage">
				{sysUsage.ram.includes("(")
					? sysUsage.ram.split("(").pop().replace(")", "")
					: "0%"}
			</span>
		</div>
		<MiniResourceGraph
			data={history.ram}
			color="var(--accent-secondary, #b197fc)"
			height={46}
		/>
		<span class="info-text">{sysUsage.ram.split(" / ")[0]} used</span>
	</div>

	<!-- GPU -->
	<div class="status-box gpu-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">videogame_asset</span>
				<span class="label">GPU</span>
			</div>
			<span class="usage">{currentGpuUsage}</span>
		</div>
		<MiniResourceGraph
			data={currentGpuHistory}
			color="var(--accent-secondary, #b197fc)"
			height={46}
		/>

		<div class="gpu-slider">
			{#if gpuList.length > 1}
				<button class="gpu-arrow left" on:click={prevGpu} aria-label="Previous GPU">
					<span class="material-icons">chevron_left</span>
				</button>
			{/if}
			<div class="gpu-name-wrap">
				{#key gpuIdx}
					<span
						class="info-text gpu-name"
						title="{currentGpu} ({sysInfo.driver})"
						in:fly={{ x: 14 * slideDir, duration: 180 }}
						out:fly={{ x: -14 * slideDir, duration: 180 }}
					>
						{currentGpu}
					</span>
				{/key}
			</div>
			{#if gpuList.length > 1}
				<button class="gpu-arrow right" on:click={nextGpu} aria-label="Next GPU">
					<span class="material-icons">chevron_right</span>
				</button>
			{/if}
		</div>

		{#if gpuList.length > 1}
			<div class="gpu-dots">
				{#each gpuList as _, i}
					<span class="gpu-dot" class:active={i === gpuIdx}></span>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.status-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 20px;
	}

	.status-box {
		background: var(--bg-base);
		padding: 16px;
		border-radius: var(--radius-md);
		border: 2px solid var(--glass-border);
		display: flex;
		flex-direction: column;
		min-width: 0;
		transition: border-color var(--transition-fast);

		&:hover {
			border-color: var(--glass-border-bright);
		}

		.box-header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 12px;

			.icon-label {
				display: flex;
				align-items: center;
				gap: 6px;
				color: var(--text-muted);

				.mini-icon {
					font-size: 16px;
				}

				.label {
					font-size: 0.75rem;
					font-weight: 800;
					text-transform: uppercase;
					letter-spacing: 0.5px;
				}
			}

			.usage {
				font-size: 0.95rem;
				font-weight: 800;
				color: var(--accent-primary);
			}
		}

		.info-text {
			display: block;
			font-size: 0.75rem;
			font-weight: 700;
			color: var(--text-muted);
			margin-top: 10px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.gpu-slider {
			display: flex;
			align-items: center;
			gap: 4px;
			margin-top: 6px;
			min-width: 0;

			.gpu-name-wrap {
				position: relative;
				flex: 1;
				min-width: 0;
				height: 22px;

				.gpu-name {
					position: absolute;
					inset: 0;
					margin-top: 0;
					line-height: 22px;
				}
			}

			.gpu-arrow {
				flex-shrink: 0;
				background: transparent;
				border: none;
				color: var(--text-muted);
				width: 20px;
				height: 22px;
				padding: 0;
				display: grid;
				place-items: center;
				cursor: pointer;
				border-radius: var(--radius-sm);
				opacity: 0;
				transition: opacity var(--transition-fast), background var(--transition-fast), color var(--transition-fast);

				.material-icons {
					font-size: 16px;
				}

				&:hover {
					background: var(--accent-glow);
					color: var(--text-main);
				}
			}
		}

		&:hover .gpu-arrow {
			opacity: 1;
		}

		.gpu-dots {
			display: flex;
			justify-content: center;
			gap: 5px;
			margin-top: 6px;

			.gpu-dot {
				width: 5px;
				height: 5px;
				border-radius: 999px;
				background: rgba(255, 255, 255, 0.15);
				transition: background var(--transition-fast), width var(--transition-fast);

				&.active {
					background: var(--accent-secondary, #b197fc);
					width: 14px;
				}
			}
		}

		.system-info {
			display: flex;
			flex-direction: column;
			gap: 4px;
			overflow: hidden;

			.os-text {
				font-size: 0.85rem;
				font-weight: 800;
				color: var(--text-main);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.kernel-text {
				font-size: 0.7rem;
				font-weight: 600;
				color: var(--text-muted);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}
	}
</style>
