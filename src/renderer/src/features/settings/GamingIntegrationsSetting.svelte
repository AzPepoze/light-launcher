<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
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

<Section
	title="Game Activity & Integrations"
	icon="sports_esports"
	help="Playtime history and Discord Rich Presence for your games."
>
	<div class="stack">
		<SettingRow
			title="Playtime & recent games"
			help="Track recent launches and accumulated playtime for the Game Manager. Running-session controls still work when this is disabled."
		>
			<Button
				variant={appSettings.TrackPlaytime !== false ? "primary" : "secondary"}
				icon={appSettings.TrackPlaytime !== false ? "history" : "history_toggle_off"}
				onclick={togglePlaytimeTracking}
			>
				Playtime Tracking: {appSettings.TrackPlaytime !== false ? "ON" : "OFF"}
			</Button>
		</SettingRow>

		<SettingRow
			title="Discord Rich Presence"
			help="Publish the active LightLauncher game and elapsed session time to Discord. The game instance owns the connection, so presence can remain active after the launcher window closes."
		>
			<Button
				variant={appSettings.DiscordRichPresence !== false ? "primary" : "secondary"}
				icon="hub"
				onclick={toggleDiscordPresence}
			>
				Discord Presence: {appSettings.DiscordRichPresence !== false ? "ON" : "OFF"}
			</Button>
		</SettingRow>

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
</Section>

<style lang="scss">
	.stack {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.hint {
		margin: 0;
		color: var(--text-dim);
		font-size: 0.78rem;
	}

	.client-id-field {
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 0.8rem;
		font-weight: 800;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.client-id-row {
		display: flex;
		gap: 10px;
	}

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
