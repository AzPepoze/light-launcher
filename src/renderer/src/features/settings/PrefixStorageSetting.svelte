<script lang="ts">
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

<div class="zone-card">
	<div class="zone-header">
		<span class="material-icons">folder_special</span>
		<h2>Wine Prefix Storage Location</h2>
	</div>

	<div class="settings-section">
		<p class="desc">
			Wine prefixes can take dozens of gigabytes. You can configure a custom folder on another drive or dedicated storage partition.
		</p>

		<div class="path-box">
			<div class="path-info">
				<span class="label">{appSettings.CustomPrefixDir ? "Custom Location:" : "Default Location:"}</span>
				<span class="path" title={appSettings.CustomPrefixDir || "~/.config/light-launcher/prefixes"}>
					{appSettings.CustomPrefixDir || "~/.config/light-launcher/prefixes"}
				</span>
			</div>

			<div class="actions">
				<button class="btn primary" onclick={handleBrowsePrefixDir}>
					<span class="material-icons mini-icon">folder_open</span>
					Change Location
				</button>
				{#if appSettings.CustomPrefixDir}
					<button class="btn secondary" onclick={handleResetDefault}>
						<span class="material-icons mini-icon">restore</span>
						Reset to Default
					</button>
				{/if}
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 12px;

		.desc {
			margin: 0 0 8px 0;
			font-size: 0.9rem;
			font-weight: 600;
			color: var(--text-muted);
		}
	}

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

			.mini-icon {
				font-size: 18px;
				margin-right: 6px;
			}
		}
	}
</style>
