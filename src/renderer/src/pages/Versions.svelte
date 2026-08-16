<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { ScanProtonVersions, OpenURL } from "@lib/api";
	import * as core from "@shared";
	import { notifications } from "@stores/notificationStore";
	import { onMount } from "svelte";

	import externalIcon from "@icons/protron_forked.png";
	import steamIcon from "@icons/steam.png";

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
		OpenURL(url);
	}
</script>

<div class="versions-container">
	<PageHeader title="Proton Versions" icon={externalIcon} isCustomIcon={true} subtitle="{protonVersions.length} installed">
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
	</PageHeader>

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
		padding: 40px 48px;
		height: 100%;
		display: flex;
		flex-direction: column;
	}


	.versions-list {
		display: flex;
		flex-direction: column;
		gap: 18px;
		overflow-y: auto;
		padding-right: 8px;
		flex: 1;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border-bright);
			border-radius: 10px;
		}
	}

	.version-card {
		display: flex;
		align-items: center;
		gap: 20px;
		padding: 20px;
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-lg);
		transition: transform var(--transition-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			background: var(--bg-surface);
			border-color: var(--accent-primary);
			transform: translateX(8px);
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
		}

		.icon {
			font-size: 1.5rem;
			background: var(--bg-base);
			border: 2px solid var(--glass-border);
			width: 52px;
			height: 52px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: var(--radius-md);

			.tool-icon {
				width: 32px;
				height: 32px;
				opacity: 0.9;
				object-fit: contain;
				filter: brightness(0) invert(1);

				:global([data-theme="light"]) & {
					filter: brightness(0);
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
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.type-badge {
			font-size: 0.7rem;
			padding: 6px 14px;
			border-radius: var(--radius-sm);
			background: var(--accent-primary);
			border: none;
			color: var(--bg-base);
			font-weight: 900;
			text-transform: uppercase;
			letter-spacing: 1px;

			&.steam {
				background: var(--accent-secondary);
				color: #000000;
			}
		}
	}


	.loading {
		text-align: center;
		color: var(--text-muted);
		margin-top: 48px;
		font-weight: 700;
	}
</style>
