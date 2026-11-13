<script lang="ts">
	import { onDestroy, onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import NumberField from "@components/shared/NumberField.svelte";
	import Section from "@components/shared/Section.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import { DOWNLOAD_NUMBER_RANGES, type DownloadNumberKey } from "./downloadsSettings";
	import { DownloadSettingsState } from "./downloadsSettingsState.svelte";

	const downloads = new DownloadSettingsState();

	onMount(() => {
		void downloads.load();
	});

	onDestroy(() => {
		downloads.flush();
	});
</script>

{#snippet numberField(key: DownloadNumberKey, label: string, help: string)}
	<NumberField
		id="download-{key}"
		{label}
		value={String(downloads.settings[key])}
		min={DOWNLOAD_NUMBER_RANGES[key].min}
		max={DOWNLOAD_NUMBER_RANGES[key].max}
		oninput={(text) => downloads.setNumber(key, text)}
	>
		<p class="help">{help}</p>
	</NumberField>
{/snippet}

<Section
	title="Downloads"
	icon="download"
	help="How game files are fetched. Changes save a moment after you stop editing."
>
	{#if downloads.loadFailed}
		<Notice tone="error" icon="error">
			<p>Could not load the download settings.</p>
		</Notice>
		<div class="retry">
			<Button variant="secondary" icon="refresh" onclick={() => downloads.load()}>Try again</Button>
		</div>
	{/if}

	<fieldset class="fields" disabled={!downloads.loaded}>
		<div class="number-grid">
			{@render numberField("connections", "Connections per file", "Parts downloaded at once for each file.")}
			{@render numberField("maxActive", "Simultaneous downloads", "Files that download at the same time.")}
			{@render numberField("hostConnections", "Connections per host", "Connections to one site across all downloads.")}
			{@render numberField("retries", "Retries", "Times a failed request is tried again.")}
			{@render numberField("speedLimitKBps", "Speed limit (KB/s)", "Total speed for all downloads. 0 means unlimited.")}
		</div>

		<SlideButton
			label="Keep archives"
			subtitle="Keep the archive file after it is installed."
			checked={downloads.settings.keepArchives}
			onChange={(enabled) => downloads.setKeepArchives(enabled)}
		/>
	</fieldset>
</Section>

<style lang="scss">
	.fields {
		display: flex;
		flex-direction: column;
		gap: 20px;
		min-width: 0;
		margin: 0;
		padding: 0;
		border: 0;
	}

	.number-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 20px;
	}

	.help {
		margin: 6px 0 0;
		color: var(--text-muted);
		font-size: 0.8rem;
	}

	.retry {
		margin-top: 12px;
	}
</style>
