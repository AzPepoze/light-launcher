<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import type { SettingsState } from "./settingsState.svelte";
	import { normalizePluginRegistryUrl } from "./pluginRegistry";

	let { settings }: { settings: SettingsState } = $props();

	const savedUrl = $derived(settings.appSettings.pluginRegistryURL ?? "");
	const dirty = $derived(normalizePluginRegistryUrl(settings.pluginRegistryUrl) !== savedUrl);

	function setUrl(event: Event): void {
		settings.setPluginRegistryUrl((event.currentTarget as HTMLInputElement).value);
	}

	/** Saves when focus leaves the row. A button inside the row keeps the draft unsaved until clicked. */
	function saveOnFocusOut(event: FocusEvent): void {
		const field = event.currentTarget as HTMLElement | null;
		const next = event.relatedTarget;
		if (field && next instanceof Node && field.contains(next)) return;
		void settings.savePluginRegistryUrl();
	}
</script>

<Section
	title="Community plugins"
	icon="public"
	help="Lists installable community plugins on the Plugins page. Nothing is downloaded until you open the Community list."
>
	<SettingRow
		title="Community plugin index"
		help="HTTPS URL of a version-1 plugins.json community index. Leave empty to disable."
	>
		<!-- Only saves the settings value; the index is fetched when the Community list opens. -->
		<div class="registry-field" onfocusout={saveOnFocusOut}>
			<input
				class="input"
				type="url"
				value={settings.pluginRegistryUrl}
				oninput={setUrl}
				placeholder="https://example.com/plugins.json"
				aria-label="Community plugin index URL"
				spellcheck="false"
				autocomplete="off"
				disabled={settings.loading}
			/>
			<Button
				size="sm"
				icon="save"
				loading={settings.savingRegistryUrl}
				disabled={settings.loading || !dirty}
				onclick={() => void settings.savePluginRegistryUrl()}
			>
				Save
			</Button>
			{#if settings.pluginRegistryUrl}
				<Button
					size="sm"
					icon="clear"
					disabled={settings.loading || settings.savingRegistryUrl}
					onclick={() => void settings.clearPluginRegistryUrl()}
				>
					Clear
				</Button>
			{/if}
		</div>
	</SettingRow>
</Section>

<style lang="scss">
	.registry-field {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 8px;
		min-width: 0;
	}

	.registry-field .input {
		width: clamp(220px, 26vw, 380px);
		min-width: 0;
		font-family: monospace;
		font-size: 0.82rem;
	}
</style>
