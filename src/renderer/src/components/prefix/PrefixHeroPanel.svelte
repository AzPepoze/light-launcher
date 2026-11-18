<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import BrowseInput from "@components/shared/BrowseInput.svelte";
	import Button from "@components/shared/Button.svelte";
	import Dropdown, { asOptions } from "@components/shared/Dropdown.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import Section from "@components/shared/Section.svelte";
	import { navigateTo } from "@stores/navigationStore";

	export let currentPrefixName: string;
	export let protonDisplayNames: string[];
	export let chosenProtonName: string;
	export let prefixPath: string;
	export let heroSize: string = "--";
	export let heroGamesLinked: string = "0 games";
	export let heroCreated: string = "Created --";

	$: protonItems = asOptions(protonDisplayNames);
	$: hasProton = protonDisplayNames.length > 0;
	$: runtimeInstalled = protonDisplayNames.includes(chosenProtonName);
	$: runtimeLabel = getRuntimeLabel(chosenProtonName, runtimeInstalled);

	function getRuntimeLabel(name: string, installed: boolean): string {
		if (installed) return name;
		if (name) return "Saved runtime not installed";
		return "No runtime selected";
	}

	function openVersions() {
		navigateTo("versions");
	}
</script>

<header class="active-header">
	<Icon name="folder" size={32} class="header-icon" />
	<div class="header-text">
		<p class="header-label">Active prefix</p>
		<h2 class="header-name">{currentPrefixName}</h2>
	</div>
	<Badge tone={runtimeInstalled ? "accent" : "warning"} icon="science">
		{runtimeLabel}
	</Badge>
</header>

<Section title="Runtime" help="Proton runs the games and tools in this prefix.">
	{#if hasProton}
		<div class="runtime-field">
			<span class="field-title">Proton version</span>
			<span class="field-help">Used for every tool and game in this prefix.</span>
			<div class="runtime-control">
				<Dropdown
					options={protonItems}
					bind:value={chosenProtonName}
					placeholder="Choose a Proton version"
				/>
			</div>
		</div>
	{:else}
		<Notice tone="warning" icon="info">
			<div class="notice-action">
				<p>No Proton installed. Install one on the Versions page.</p>
				<Button size="sm" icon="open_in_new" onclick={openVersions}>Open Versions</Button>
			</div>
		</Notice>
	{/if}
</Section>

<Section title="Location" help="The folder that holds this prefix's files.">
	<div class="location-field">
		<label class="field-label" for="prefix-path">Prefix folder</label>
		<BrowseInput
			id="prefix-path"
			bind:value={prefixPath}
			placeholder="Path to Wine prefix"
			type="folder"
		/>
	</div>
	<div class="facts">
		<Badge icon="hard_drive">{heroSize}</Badge>
		<Badge icon="sports_esports">{heroGamesLinked}</Badge>
		<Badge icon="event">{heroCreated}</Badge>
	</div>
</Section>

<style lang="scss">
	.active-header {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 18px 24px;
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-lg);
		flex-wrap: wrap;

		:global(.header-icon) {
			color: var(--accent-primary);
			flex-shrink: 0;
		}
	}

	.header-text {
		flex: 1;
		min-width: 0;
	}

	.header-label {
		margin: 0 0 2px;
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	.header-name {
		margin: 0;
		font-size: 1.3rem;
		font-weight: 800;
		color: var(--text-main);
		line-height: 1.2;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.runtime-field {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.field-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text-main);
	}

	.field-help {
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.runtime-control {
		width: min(360px, 100%);
		margin-top: 8px;
	}

	.notice-action {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;

		p {
			margin: 0;
		}
	}

	.location-field {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.field-label {
		font-size: 0.85rem;
		font-weight: 700;
		color: var(--text-main);
	}

	.facts {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
</style>
