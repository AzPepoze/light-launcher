<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Button from "@components/shared/Button.svelte";
	import Badge from "@components/shared/Badge.svelte";
	import { ScanProtonVersions, OpenURL } from "@lib/api";
	import * as core from "@shared";
	import { notifications } from "@stores/notificationStore";
	import { onMount } from "svelte";
	import { createLogger } from "@lib/logger";

	import externalIcon from "@icons/protron_forked.png";
	import steamIcon from "@icons/steam.png";

	const log = createLogger("Versions");

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
			log.error("Failed to scan versions", err);
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
		<Button
			variant="secondary"
			icon="travel_explore"
			title="ProtonDB: check game compatibility reports (opens in browser)"
			onclick={() => openExternal("https://protondb.com")}
		>
			ProtonDB
		</Button>
		<Button
			variant="secondary"
			icon="extension"
			title="ProtonPlus: manage Proton versions (opens GitHub)"
			onclick={() =>
				openExternal("https://github.com/Vysp3r/ProtonPlus")}
		>
			ProtonPlus
		</Button>
		<Button
			variant="secondary"
			icon="download"
			title="ProtonUp-Qt: install Proton versions (opens GitHub)"
			onclick={() =>
				openExternal(
					"https://github.com/DavidoTek/ProtonUp-Qt",
				)}
		>
			ProtonUp
		</Button>
	</PageHeader>

	<p class="legend">
		<Badge tone="accent">Steam</Badge> found in Steam's compatibility folders
		<Badge>External</Badge> found in other tool folders
	</p>

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
					<Badge tone={tool.IsSteam ? "accent" : "neutral"}>
						{tool.IsSteam ? "Steam" : "External"}
					</Badge>
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

	@media (max-width: 600px) {
		.versions-container {
			padding: 24px 0;
		}

		.versions-container :global(.header-actions) {
			flex-wrap: wrap;
			flex-shrink: 1;
			min-width: 0;
			max-width: 100%;
		}
	}

	.legend {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin: 0 0 18px;
		font-size: 0.8rem;
		color: var(--text-muted);
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
			min-width: 0;
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

	}

	@media (max-width: 600px) {
		.version-card {
			gap: 12px;
			padding: 14px;
		}

		.version-card .path {
			white-space: normal;
			overflow-wrap: anywhere;
		}
	}

	.loading {
		text-align: center;
		color: var(--text-muted);
		margin-top: 48px;
		font-weight: 700;
	}
</style>
