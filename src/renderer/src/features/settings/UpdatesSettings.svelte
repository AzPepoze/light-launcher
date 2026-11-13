<script lang="ts">
	import { onDestroy, onMount } from "svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import type { LibrarySource } from "@lib/features/types";
	import type { SettingsState } from "./settingsState.svelte";
	import {
		INTERVAL_OPTIONS,
		RECHECK_OPTIONS,
		SOURCE_MODE_OPTIONS,
		globalUpdateChecks,
		hoursFromOption,
		optionsWithHours,
		sourceHelp,
		sourceModeFromOption,
		sourceView,
		withGlobalUpdateChecks,
		withSourceMode,
		withSourceRecheck
	} from "./updateChecks";
	import { updateCheckSources, type UpdateSource } from "./updateSources";

	let {
		settings,
		sources
	}: {
		settings: SettingsState;
		/** The library sources, already loaded by the page. Missing means load the overview here. */
		sources?: LibrarySource[];
	} = $props();

	const checks = $derived(settings.appSettings.updateChecks);
	const global = $derived(globalUpdateChecks(checks));
	const intervalOptions = $derived(optionsWithHours(INTERVAL_OPTIONS, global.intervalHours));
	const recheckOptions = $derived(optionsWithHours(RECHECK_OPTIONS, global.recheckHours));
	const rows = $derived(updateCheckSources(sources ?? settings.sources));

	let loadingSources = $state(false);

	onMount(() => {
		if (sources) return;
		loadingSources = true;
		void settings.loadSources().finally(() => (loadingSources = false));
	});

	onDestroy(() => {
		settings.flushUpdateChecks();
	});
</script>

{#snippet sourceRow(source: UpdateSource)}
	{@const view = sourceView(checks, source.id)}
	{@const sourceRecheckOptions = optionsWithHours(RECHECK_OPTIONS, view.recheckHours)}
	<div class="source-row">
		<div class="source-text">
			<span class="source-name">
				<span class="source-icon"><Icon name={source.icon} size={18} /></span>
				{source.label}
			</span>
			<span class="source-help">{sourceHelp(view.mode)}</span>
		</div>
		<div class="source-controls">
			<Dropdown
				options={SOURCE_MODE_OPTIONS}
				value={view.mode}
				disabled={!global.enabled}
				onChange={(value) =>
					settings.setUpdateChecks(withSourceMode(checks, source.id, sourceModeFromOption(value)))}
			/>
			{#if view.mode === "custom"}
				<Dropdown
					options={sourceRecheckOptions}
					value={String(view.recheckHours)}
					disabled={!global.enabled}
					onChange={(value) =>
						settings.setUpdateChecks(
							withSourceRecheck(checks, source.id, hoursFromOption(value, view.recheckHours))
						)}
				/>
			{/if}
		</div>
	</div>
{/snippet}

<Section
	title="Updates"
	icon="update"
	help="When the launcher checks your games for new versions."
>
	<SlideButton
		label="Check for updates in the background"
		subtitle="Checks run on their own. Check now always works."
		checked={global.enabled}
		onChange={(enabled) => settings.setUpdateChecks(withGlobalUpdateChecks(checks, { enabled }))}
	/>

	<SlideButton
		label="Check once after launch"
		subtitle="A quiet check a short time after the app starts."
		checked={global.onStartup}
		onChange={(onStartup) => settings.setUpdateChecks(withGlobalUpdateChecks(checks, { onStartup }))}
	/>

	<SettingRow title="Check every" help="How often background checks run. Never turns them off.">
		<Dropdown
			options={intervalOptions}
			value={String(global.intervalHours)}
			onChange={(value) =>
				settings.setUpdateChecks(
					withGlobalUpdateChecks(checks, { intervalHours: hoursFromOption(value, global.intervalHours) })
				)}
		/>
	</SettingRow>

	<SettingRow title="Skip games checked within" help="Games checked this recently are left out of background checks.">
		<Dropdown
			options={recheckOptions}
			value={String(global.recheckHours)}
			onChange={(value) =>
				settings.setUpdateChecks(
					withGlobalUpdateChecks(checks, { recheckHours: hoursFromOption(value, global.recheckHours) })
				)}
		/>
	</SettingRow>

	<div class="sources">
		<h3>Sources</h3>

		{#if !global.enabled}
			<p class="note">Background checks are off, so no source checks in the background.</p>
		{/if}

		{#each rows as source (source.id)}
			{@render sourceRow(source)}
		{:else}
			{#if loadingSources}
				<LoadingState label="Loading sources..." />
			{:else}
				<p class="note">No plugin source can check for updates.</p>
			{/if}
		{/each}

		<Notice tone="info" icon="info">
			<p>Check now on the Updates page ignores these limits.</p>
		</Notice>
	</div>
</Section>

<style lang="scss">
	.sources {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-top: 8px;

		h3 {
			margin: 8px 0 4px;
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 700;
		}
	}

	.source-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding: 16px 0;

		& + .source-row {
			border-top: 2px solid rgba(255, 255, 255, 0.05);
		}
	}

	.source-text {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.source-name {
		display: flex;
		align-items: center;
		gap: 8px;
		color: var(--text-main);
		font-size: 0.95rem;
		font-weight: 600;
	}

	.source-icon {
		display: inline-flex;
		color: var(--text-muted);
	}

	.source-help {
		color: var(--text-muted);
		font-size: 0.8rem;
	}

	.source-controls {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 12px;
	}

	.note {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	p {
		margin: 0;
	}
</style>
