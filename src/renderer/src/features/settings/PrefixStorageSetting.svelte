<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import type { SettingsState } from "./settingsState.svelte";

	const DEFAULT_PREFIX_DIR = "~/.config/light-launcher/prefixes";

	let { settings }: { settings: SettingsState } = $props();

	const savedDir = $derived(settings.appSettings.CustomPrefixDir ?? "");
	let draft = $state("");
	const dirty = $derived(draft.trim() !== savedDir);

	// Keeps the draft in step when the saved folder changes, for example after Browse.
	$effect(() => {
		draft = savedDir;
	});

	function handleBrowsePrefixDir() {
		return settings.browsePrefixDirectory();
	}

	function handleResetDefault() {
		return settings.resetPrefixDirectory();
	}
</script>

<Section
	title="Wine Prefix Storage Location"
	icon="folder_special"
	help="Wine prefixes can take dozens of gigabytes. You can configure a custom folder on another drive or dedicated storage partition."
>
	<div class="prefix-setting">
		<SettingRow
			title="Wine prefix location"
			help="Wine prefixes can take dozens of gigabytes. Use a folder on another drive or partition."
		>
			<div class="prefix-field">
				<input
					class="input"
					type="text"
					bind:value={draft}
					placeholder={DEFAULT_PREFIX_DIR}
					aria-label="Wine prefix folder"
					spellcheck="false"
					autocomplete="off"
					disabled={settings.loading}
				/>
				<Button
					size="sm"
					icon="save"
					loading={settings.savingPrefixDir}
					disabled={settings.loading || !dirty}
					onclick={() => void settings.savePrefixDirectory(draft)}
				>
					Save
				</Button>
				<IconButton
					icon="folder_open"
					label="Browse folder"
					disabled={settings.loading || settings.savingPrefixDir}
					onclick={handleBrowsePrefixDir}
				/>
				{#if savedDir}
					<Button
						size="sm"
						icon="restore"
						disabled={settings.loading || settings.savingPrefixDir}
						onclick={handleResetDefault}
					>
						Reset to Default
					</Button>
				{/if}
			</div>
		</SettingRow>
	</div>
</Section>

<style lang="scss">
	.prefix-field {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 8px;
		min-width: 0;

		.input {
			width: clamp(220px, 26vw, 380px);
			min-width: 0;
			font-family: monospace;
			font-size: 0.82rem;
		}
	}

	@media (max-width: 599px) {
		.prefix-setting :global(.setting-row) {
			flex-direction: column;
			align-items: stretch;
		}

		.prefix-setting :global(.control),
		.prefix-field,
		.prefix-field .input {
			width: 100%;
		}

		.prefix-field {
			justify-content: flex-start;
		}
	}
</style>
