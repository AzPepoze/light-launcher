<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import type { SettingsState } from "./settingsState.svelte";

	let { settings } = $props<{ settings: SettingsState }>();
	const appSettings = $derived(settings.appSettings);
	const savingDiscordId = $derived(settings.savingDiscordId);

	function togglePlaytimeTracking() {
		return settings.togglePlaytimeTracking();
	}

	function toggleDiscordPresence() {
		return settings.toggleDiscordPresence();
	}

	function saveDiscordClientId() {
		return settings.saveDiscordClientId();
	}
</script>

<div class="zone-card zone-card--flat">
	<div class="zone-header">
		<Icon name="sports_esports" size={24} class="zone-icon" />
		<h2>Game Activity & Integrations</h2>
	</div>

	<div class="settings-section">
		<h3>Playtime & Recent Games</h3>
		<p class="desc">
			Track recent launches and accumulated playtime for the Game Manager. Running-session controls still work when this is disabled.
		</p>
		<Button
			variant={appSettings.TrackPlaytime !== false ? "primary" : "secondary"}
			icon={appSettings.TrackPlaytime !== false ? "history" : "history_toggle_off"}
			onclick={togglePlaytimeTracking}
		>
			Playtime Tracking: {appSettings.TrackPlaytime !== false ? "ON" : "OFF"}
		</Button>
	</div>

	<div class="divider"></div>

	<div class="settings-section">
		<h3>Discord Rich Presence</h3>
		<p class="desc">
			Publish the active LightLauncher game and elapsed session time to Discord. The game instance owns the connection, so presence can remain active after the launcher window closes.
		</p>
		<Button
			variant={appSettings.DiscordRichPresence !== false ? "primary" : "secondary"}
			icon="hub"
			onclick={toggleDiscordPresence}
		>
			Discord Presence: {appSettings.DiscordRichPresence !== false ? "ON" : "OFF"}
		</Button>

		<label class="client-id-field">
			<span>Discord Application ID</span>
			<div class="client-id-row">
				<input
					class="input"
					type="text"
					value={settings.discordClientId}
					oninput={(event) => settings.setDiscordClientId(event.currentTarget.value)}
					placeholder="Discord application client ID"
					autocomplete="off"
					spellcheck="false"
				/>
				<Button
					variant="secondary"
					icon="save"
					loading={savingDiscordId}
					onclick={saveDiscordClientId}
				>
					{savingDiscordId ? "Saving..." : "Save"}
				</Button>
			</div>
		</label>
		<p class="hint">You can also set <code>LIGHT_LAUNCHER_DISCORD_CLIENT_ID</code>. If Discord is closed or unavailable, game launch continues normally.</p>
	</div>
</div>

<style lang="scss">
	.settings-section { display: flex; flex-direction: column; gap: 10px; }
	.settings-section h3 { margin: 0; font-size: 1.05rem; text-transform: uppercase; }
	.desc, .hint { margin: 0 0 8px; color: var(--text-muted); font-size: 0.88rem; line-height: 1.55; }
	.hint { margin-top: 2px; font-size: 0.78rem; color: var(--text-dim); }
	.divider { height: 2px; background: rgba(255, 255, 255, 0.05); border-radius: var(--radius-pill); }

	:global(.zone-icon) {
		color: var(--accent-primary);
	}

	.client-id-field {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin-top: 8px;
		font-size: 0.8rem;
		font-weight: 800;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.client-id-row { display: flex; gap: 10px; }
	.client-id-row .input {
		flex: 1;
		min-width: 0;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 10px 14px;
		color: var(--text-main);
		outline: none;
		font-family: monospace;
		text-transform: none;
		letter-spacing: normal;
		&:focus { border-color: var(--accent-primary); }
	}
</style>
