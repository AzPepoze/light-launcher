<script lang="ts">
	export let systemStatus: {
		hasGamescope: boolean;
		hasMangoHud: boolean;
		hasGameMode: boolean;
		hasLosslessDll: boolean;
		hasVulkanInfo: boolean;
	};
</script>

<div class="section-container glass">
	<h3>System Dependencies</h3>
	<div class="system-status-grid">
		<div class="status-item" class:ok={systemStatus.hasGamescope}>
			<span class="dot"></span>
			<span class="label">Gamescope</span>
			<span class="value">
				{systemStatus.hasGamescope ? "Available" : "Not Found"}
			</span>
		</div>
		<div class="status-item" class:ok={systemStatus.hasMangoHud}>
			<span class="dot"></span>
			<span class="label">MangoHud</span>
			<span class="value">
				{systemStatus.hasMangoHud ? "Available" : "Not Found"}
			</span>
		</div>
		<div class="status-item" class:ok={systemStatus.hasGameMode}>
			<span class="dot"></span>
			<span class="label">GameMode</span>
			<span class="value">
				{systemStatus.hasGameMode ? "Available" : "Not Found"}
			</span>
		</div>
		<div class="status-item" class:ok={systemStatus.hasVulkanInfo}>
			<span class="dot"></span>
			<span class="label">Vulkan-Tools</span>
			<span class="value">
				{systemStatus.hasVulkanInfo ? "Available" : "Missing"}
			</span>
		</div>
		<div class="status-item" class:ok={systemStatus.hasLosslessDll}>
			<span class="dot"></span>
			<span class="label">Lossless.dll</span>
			<span class="value">
				{systemStatus.hasLosslessDll ? "Found" : "Not Found"}
			</span>
		</div>
	</div>
	
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

	.system-status-grid {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
	}

	.status-item {
		display: flex;
		align-items: center;
		gap: 12px;
		font-size: 0.9rem;
		padding: 10px 20px;
		background: var(--bg-base);
		border: 2px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-pill);
		color: var(--text-muted);
		transition: transform var(--transition-spring), border-color var(--transition-fast);

		&:hover {
			transform: translateY(-2px);
			border-color: rgba(255, 255, 255, 0.1);
		}

		.dot {
			width: 10px;
			height: 10px;
			border-radius: 50%;
			background: var(--danger);
			box-shadow: 0 0 10px var(--danger);
		}
		.label {
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			font-size: 0.85rem;
			letter-spacing: 0.5px;
		}
		.value {
			font-size: 0.8rem;
			font-weight: 700;
			opacity: 0.8;
		}

		&.ok {
			.dot {
				background: var(--accent-secondary);
				box-shadow: 0 0 10px var(--accent-secondary);
			}
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
