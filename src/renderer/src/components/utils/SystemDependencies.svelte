<script lang="ts">
	export let systemStatus: {
		hasGamescope: boolean;
		hasMangoHud: boolean;
		hasGameMode: boolean;
		hasLosslessDll: boolean;
		hasVulkanInfo: boolean;
	};

	$: items = [
		{
			name: "Gamescope",
			ok: systemStatus.hasGamescope,
			statusText: systemStatus.hasGamescope ? "Available" : "Not Found",
			description: "Provides custom resolution sandbox, upscaling, and HDR support."
		},
		{
			name: "MangoHud",
			ok: systemStatus.hasMangoHud,
			statusText: systemStatus.hasMangoHud ? "Available" : "Not Found",
			description: "Monitors FPS, frame times, CPU/GPU temperatures, and resource usage."
		},
		{
			name: "GameMode",
			ok: systemStatus.hasGameMode,
			statusText: systemStatus.hasGameMode ? "Available" : "Not Found",
			description: "Optimizes CPU governor and scheduler settings for maximum performance."
		},
		{
			name: "Vulkan-Tools",
			ok: systemStatus.hasVulkanInfo,
			statusText: systemStatus.hasVulkanInfo ? "Available" : "Missing",
			description: "Enables GPU auto-detection and Vulkan capabilities verification."
		},
		{
			name: "Lossless.dll",
			ok: systemStatus.hasLosslessDll,
			statusText: systemStatus.hasLosslessDll ? "Found" : "Not Found",
			description: "Lossless Scaling binary libraries required for LSFG-VK frame generation."
		}
	];
</script>

<div class="section-container glass">
	<h3>System Dependencies</h3>

	<table class="status-table">
		<thead>
			<tr>
				<th>Dependency</th>
				<th>Status</th>
				<th>Description</th>
			</tr>
		</thead>
		<tbody>
			{#each items as item}
				<tr>
					<td><strong>{item.name}</strong></td>
					<td>
						<span class="status-badge" class:ok={item.ok} title={item.statusText}>
							{item.statusText}
						</span>
					</td>
					<td class="desc">{item.description}</td>
				</tr>
			{/each}
		</tbody>
	</table>

	{#if !systemStatus.hasVulkanInfo}
		<p class="warning-text important">
			<span class="material-icons icon">warning</span>
			Vulkan-Tools (vulkaninfo) is missing. This is highly recommended for accurate GPU detection and LSFG compatibility.
		</p>
	{/if}

	{#if !systemStatus.hasGamescope || !systemStatus.hasMangoHud || !systemStatus.hasGameMode || !systemStatus.hasLosslessDll}
		<p class="warning-text">
			Some features may not work until you install these tools.
			LSFG-VK requires Lossless Scaling installed via Steam.
		</p>
	{/if}
</div>

<style lang="scss">
	.section-container {
		padding: 24px 28px;
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		background: var(--bg-surface);
		margin-bottom: 32px;

		h3 {
			margin: 0 0 20px 0;
			font-size: 1.15rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 1px;
		}
	}

	.status-table {
		width: 100%;
		border-collapse: collapse;
		margin-top: 12px;
		text-align: left;

		th, td {
			padding: 14px 18px;
			font-size: 0.9rem;
			border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		}

		th {
			font-weight: 800;
			text-transform: uppercase;
			color: var(--text-muted);
			font-size: 0.8rem;
			letter-spacing: 0.5px;
			background: rgba(0, 0, 0, 0.15);
		}

		tr {
			transition: background-color var(--transition-fast);
			&:hover {
				background: rgba(255, 255, 255, 0.02);
			}
		}

		td strong {
			color: var(--text-main);
			font-weight: 700;
		}

		td.desc {
			color: var(--text-muted);
			font-size: 0.85rem;
			font-weight: 500;
		}
	}

	.status-badge {
		display: inline-block;
		max-width: 100%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		vertical-align: middle;
		text-align: center;
		font-size: 0.8rem;
		font-weight: 700;
		color: var(--danger);
		padding: 6px 14px;
		background: rgba(255, 59, 48, 0.1);
		border: 1px solid rgba(255, 59, 48, 0.2);
		border-radius: var(--radius-pill);

		&.ok {
			color: var(--success);
			background: rgba(52, 199, 89, 0.1);
			border-color: rgba(52, 199, 89, 0.2);
		}
	}

	.warning-text {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-muted);
		margin: 16px 0 0 0;

		&.important {
			color: var(--danger);
			background: rgba(255, 74, 122, 0.1);
			border: 2px solid rgba(255, 74, 122, 0.15);
			padding: 12px 18px;
			border-radius: var(--radius-md);
			font-style: normal;
			display: flex;
			align-items: center;
			gap: 10px;

			.icon {
				font-size: 1.3rem;
			}
		}
	}
</style>
