<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Section from "@components/shared/Section.svelte";
	import type { SettingsState } from "./settingsState.svelte";

	let { settings } = $props<{ settings: SettingsState }>();
	const appSettings = $derived(settings.appSettings);

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
	<div class="path-box">
		<div class="path-info">
			<span class="label">{appSettings.CustomPrefixDir ? "Custom Location:" : "Default Location:"}</span>
			<span class="path" title={appSettings.CustomPrefixDir || "~/.config/light-launcher/prefixes"}>
				{appSettings.CustomPrefixDir || "~/.config/light-launcher/prefixes"}
			</span>
		</div>

		<div class="actions">
			<Button variant="primary" icon="folder_open" onclick={handleBrowsePrefixDir}>
				Change Location
			</Button>
			{#if appSettings.CustomPrefixDir}
				<Button variant="secondary" icon="restore" onclick={handleResetDefault}>
					Reset to Default
				</Button>
			{/if}
		</div>
	</div>
</Section>

<style lang="scss">
	.path-box {
		display: flex;
		flex-direction: column;
		gap: 16px;
		background: var(--bg-base);
		border: 2px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-md);
		padding: 16px 20px;

		.path-info {
			display: flex;
			flex-direction: column;
			gap: 4px;

			.label {
				font-size: 0.75rem;
				font-weight: 800;
				color: var(--text-dim);
				text-transform: uppercase;
				letter-spacing: 0.5px;
			}

			.path {
				font-size: 0.95rem;
				font-weight: 700;
				color: var(--accent-secondary, var(--text-main));
				font-family: monospace;
				word-break: break-all;
			}
		}

		.actions {
			display: flex;
			gap: 12px;
			align-items: center;
		}
	}
</style>
