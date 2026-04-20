<script lang="ts">
	import { onMount } from "svelte";
	import { ScanProtonVersions } from "@bindings/light-launcher-wails/backend/app";
	import { Browser } from "@wailsio/runtime";
	import * as core from "@bindings/light-launcher/pkg/core/models";
	import { notifications } from "@stores/notificationStore";

	import steamIcon from "@icons/steam.png";
	import externalIcon from "@icons/protron_forked.png";

	let protonVersions: core.ProtonTool[] = [];
	let isLoading = true;

	onMount(async () => {
		loadInstalledVersions();
	});

	async function loadInstalledVersions() {
		isLoading = true;
		try {
			protonVersions = await ScanProtonVersions();
		} catch (err) {
			console.error(err);
			notifications.error("Failed to scan versions");
		} finally {
			isLoading = false;
		}
	}

	function openExternal(url: string) {
		Browser.OpenURL(url);
	}
</script>

<div class="versions-container">
	<div class="header">
		<div class="title-group">
			<h1 class="page-title">Proton Versions</h1>
			<div class="count-badge">{protonVersions.length} Installed</div>
		</div>

		<div class="actions">
			<button
				class="btn secondary"
				on:click={() => openExternal("https://protondb.com")}
			>
				ProtonDB
			</button>
			<button
				class="btn secondary"
				on:click={() =>
					openExternal("https://github.com/Vysp3r/ProtonPlus")}
			>
				ProtonPlus
			</button>
			<button
				class="btn secondary"
				on:click={() =>
					openExternal(
						"https://github.com/DavidoTek/ProtonUp-Qt",
					)}
			>
				ProtonUp
			</button>
		</div>
	</div>

	<div class="versions-list">
		{#if isLoading}
			<div class="loading">Scanning...</div>
		{:else}
			{#each protonVersions as tool}
				<div class="version-card">
					<div class="icon">
						<img
							src={tool.IsSteam ? steamIcon : externalIcon}
							alt="tool"
							class="tool-icon"
						/>
					</div>
					<div class="info">
						<div class="name">{tool.Name}</div>
						<div class="path" title={tool.Path}>
							{tool.Path}
						</div>
					</div>
					<div class="type-badge" class:steam={tool.IsSteam}>
						{tool.IsSteam ? "Steam" : "External"}
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>

<style lang="scss">
	.versions-container {
		padding: 32px;
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 32px;
	}

	.title-group {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.actions {
		display: flex;
		gap: 12px;
	}

	.page-title {
		font-size: 2rem;
		font-weight: 800;
		color: var(--text-main);
		margin: 0;
	}

	.count-badge {
		background: var(--glass-surface);
		padding: 4px 12px;
		border-radius: 20px;
		font-size: 0.8rem;
		color: var(--text-dim);
		border: 1px solid var(--glass-border);
	}

	.versions-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
		overflow-y: auto;
		padding-right: 8px;
		flex: 1;
	}

	.version-card {
		display: flex;
		align-items: center;
		gap: 20px;
		padding: 20px;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		transition: all 0.2s;

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--accent-primary);
			transform: translateX(4px);
		}

		.icon {
			font-size: 1.5rem;
			background: var(--glass-surface);
			border: 1px solid var(--glass-border);
			width: 52px;
			height: 52px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 12px;
			box-shadow: 0 4px 10px rgba(0, 0, 0, 0.05);

			.tool-icon {
				width: 32px;
				height: 32px;
				opacity: 0.9;
				object-fit: contain;

				/* Use dynamic filter for dark mode if needed, but in light mode keep original */
				:global([data-theme="dark"]) & {
					filter: brightness(0) invert(1);
				}
			}
		}

		.info {
			flex: 1;
			overflow: hidden;
		}

		.name {
			font-size: 1.15rem;
			font-weight: 800;
			color: var(--text-main);
			margin-bottom: 4px;
			letter-spacing: -0.3px;
		}

		.path {
			font-size: 0.8rem;
			color: var(--text-dim);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			opacity: 0.8;
		}

		.type-badge {
			font-size: 0.7rem;
			padding: 6px 14px;
			border-radius: 8px;
			background: var(--glass-hover);
			border: 1px solid var(--glass-border);
			color: var(--text-muted);
			font-weight: 800;
			text-transform: uppercase;
			letter-spacing: 1px;

			&.steam {
				background: rgba(14, 165, 233, 0.1);
				color: #0284c7; /* Solid blue for readability in both modes */
				border: 1px solid rgba(14, 165, 233, 0.3);

				:global([data-theme="dark"]) & {
					color: #38bdf8;
					background: rgba(14, 165, 233, 0.2);
				}
			}
		}
	}

	.loading {
		text-align: center;
		color: var(--text-dim);
		margin-top: 48px;
	}
</style>
