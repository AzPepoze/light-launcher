<script lang="ts">
	export let name: string;
	export let isInstalled: boolean;
	export let isInstalling: boolean = false;
	export let isUninstalling: boolean = false;
	export let progressMessage: string = "";
	export let progressPercent: number = 0;
	export let description: string[] = [];
	export let note: string = "";
	export let icon: string;
	export let onInstall: () => Promise<void>;
	export let onUninstall: () => Promise<void>;
</script>

<div class="util-card glass">
	<div class="util-header">
		<div class="icon-bg">
			<img src={icon} alt={name} class="util-logo" />
		</div>
		<div class="title-area">
			<h3>{name}</h3>
			<span class="badge" class:installed={isInstalled}>
				{isInstalled ? "Installed" : "Not Installed"}
			</span>
		</div>
	</div>

	<div class="description">
		{#each description as paragraph}
			<p>{@html paragraph}</p>
		{/each}
		{#if note}
			<p class="note">{note}</p>
		{/if}
	</div>

	<div class="action-area">
		{#if isInstalled}
			<button
				class="btn danger"
				on:click={onUninstall}
				disabled={isUninstalling}
			>
				{isUninstalling ? "Removing..." : "Remove Utility"}
			</button>
		{:else}
			<div class="install-controls">
				<button
					class="btn primary"
					on:click={onInstall}
					disabled={isInstalling}
				>
					{isInstalling ? "Installing..." : `Install ${name}`}
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
			border-color: var(--accent-secondary);
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

			.util-logo {
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
			background: rgba(52, 199, 89, 0.15);
			color: var(--success);
			border: 1px solid rgba(52, 199, 89, 0.25);
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
		:global(strong) {
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
		}
	}
</style>
