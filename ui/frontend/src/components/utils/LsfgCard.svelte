<script lang="ts">
	export let status: { isLsfgInstalled: boolean; lsfgVersion: string };
	export let isInstalling: boolean;
	export let isUninstalling: boolean;
	export let progressMessage: string;
	export let progressPercent: number;
	export let handleInstall: () => Promise<void>;
	export let handleUninstall: () => Promise<void>;
	export let lsfgPng: string;
</script>

<div class="util-card glass">
	<div class="util-header">
		<div class="icon-bg">
			<img src={lsfgPng} alt="lsfg" class="lsfg-logo" />
		</div>
		<div class="title-area">
			<h3>LSFG-VK</h3>
			<span class="badge" class:installed={status.isLsfgInstalled}>
				{status.isLsfgInstalled ? "Installed" : "Not Installed"}
			</span>
		</div>
	</div>

	<div class="description">
		<p>
			Lossless Scaling is a Windows-exclusive program featuring various algorithms for scaling and interpolating programs.
		</p>
		<p>
			<strong>lsfg-vk</strong> is a Vulkan layer that hooks into Vulkan applications and generates additional frames using Lossless Scaling's frame generation algorithm.
		</p>
		<p class="note">
			Note: Requires Lossless Scaling downloaded on Steam.
		</p>
	</div>

	<div class="action-area">
		{#if status.isLsfgInstalled}
			<button
				class="btn danger"
				on:click={handleUninstall}
				disabled={isUninstalling}
			>
				{isUninstalling ? "Removing..." : "Remove Utility"}
			</button>
		{:else}
			<div class="install-controls">
				<button
					class="btn primary"
					on:click={handleInstall}
					disabled={isInstalling}
				>
					{isInstalling ? "Installing..." : "Install LSFG-VK"}
				</button>
				{#if isInstalling}
					<div class="install-progress-area">
						<div class="progress-header">
							<span class="msg">{progressMessage}</span>
							<span class="pct">{progressPercent}%</span>
						</div>
						<div class="progress-bar-container">
							<div
								class="progress-fill"
								style="width: {progressPercent}%"
							></div>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.util-card {
		padding: 24px;
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		background: var(--bg-surface);
		display: flex;
		flex-direction: column;
		gap: 20px;
		transition: transform var(--transition-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			border-color: var(--accent-primary);
			transform: scale(1.01);
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
		}
	}

	.util-header {
		display: flex;
		align-items: center;
		gap: 20px;

		.icon-bg {
			width: 56px;
			height: 56px;
			background: var(--bg-base);
			border: 2px solid rgba(255, 255, 255, 0.05);
			border-radius: var(--radius-md);
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 1.75rem;

			.lsfg-logo {
				width: 32px;
				height: 32px;
				opacity: 0.9;
				object-fit: contain;
				filter: brightness(0) invert(1);
			}
		}

		.title-area {
			display: flex;
			flex-direction: column;
			gap: 6px;

			h3 {
				margin: 0;
				font-size: 1.25rem;
				font-weight: 800;
				text-transform: uppercase;
				letter-spacing: 0.5px;
			}
		}
	}

	.badge {
		display: inline-block;
		padding: 4px 10px;
		border-radius: var(--radius-pill);
		font-size: 0.7rem;
		font-weight: 900;
		text-transform: uppercase;
		background: var(--bg-elevated);
		color: var(--text-muted);
		width: max-content;

		&.installed {
			background: var(--accent-secondary);
			color: #000000;
		}
	}

	.description {
		font-size: 0.9rem;
		color: var(--text-muted);
		line-height: 1.6;
		margin: 0;
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 12px;

		p {
			margin: 0;
		}
		strong {
			color: var(--text-main);
			font-weight: 800;
		}
		.note {
			font-size: 0.8rem;
			color: var(--accent-primary);
			font-style: italic;
			padding-top: 12px;
			border-top: 2px solid rgba(255, 255, 255, 0.05);
		}
	}

	.install-controls {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.install-progress-area {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px;
		background: var(--bg-base);
		border-radius: var(--radius-md);
		border: 2px solid rgba(255, 255, 255, 0.05);
	}

	.progress-header {
		display: flex;
		justify-content: space-between;
		font-size: 0.85rem;
		font-weight: 800;

		.msg {
			color: var(--text-muted);
		}
		.pct {
			color: var(--accent-primary);
		}
	}

	.progress-bar-container {
		height: 8px;
		background: var(--bg-input);
		border-radius: var(--radius-pill);
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--accent-primary);
		transition: width 0.3s ease;
		box-shadow: 0 0 10px var(--accent-primary);
	}

	.action-area {
		margin-top: 8px;

		.btn {
			width: 100%;
			padding: 12px 24px;
			border-radius: var(--radius-pill);
			font-weight: 800;
			cursor: pointer;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);

			&.primary {
				background: var(--accent-primary);
				color: #ffffff;
				border: none;
				box-shadow: 0 4px 12px var(--accent-glow);

				&:hover {
					background: var(--accent-hover);
					transform: scale(1.03);
				}

				&:active {
					transform: scale(0.97);
				}
			}

			&.danger {
				background: var(--danger);
				color: #ffffff;
				border: none;

				&:hover {
					background: #ff2e63;
					transform: scale(1.03);
				}

				&:active {
					transform: scale(0.97);
				}
			}
		}
	}
</style>
