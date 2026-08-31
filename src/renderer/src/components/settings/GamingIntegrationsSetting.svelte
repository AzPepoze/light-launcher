<script lang="ts">
	import * as service from "@lib/settingsService";
	import type { AppSettings } from "@shared";

	export let appSettings: AppSettings;
	export let onRefresh: () => Promise<void> | void = () => {};

	let savingDiscordId = false;

	async function togglePlaytimeTracking() {
		appSettings = {
			...appSettings,
			TrackPlaytime: appSettings.TrackPlaytime === false
		};
		await service.saveAppSettings(
			appSettings,
			appSettings.TrackPlaytime === false ? "Playtime tracking disabled" : "Playtime tracking enabled"
		);
		await onRefresh();
	}

	async function toggleDiscordPresence() {
		appSettings = {
			...appSettings,
			DiscordRichPresence: appSettings.DiscordRichPresence === false
		};
		await service.saveAppSettings(
			appSettings,
			appSettings.DiscordRichPresence === false ? "Discord Rich Presence disabled" : "Discord Rich Presence enabled"
		);
		await onRefresh();
	}

	async function saveDiscordClientId() {
		savingDiscordId = true;
		try {
			appSettings = { ...appSettings, DiscordClientId: appSettings.DiscordClientId?.trim() || "" };
			await service.saveAppSettings(appSettings, "Discord Application ID saved");
			await onRefresh();
		} finally {
			savingDiscordId = false;
		}
	}
</script>

<div class="zone-card">
	<div class="zone-header">
		<span class="material-icons">sports_esports</span>
		<h2>Game Activity & Integrations</h2>
	</div>

	<div class="settings-section">
		<h3>Playtime & Recent Games</h3>
		<p class="desc">
			Track recent launches and accumulated playtime for the Game Manager. Running-session controls still work when this is disabled.
		</p>
		<button
			class="btn {appSettings.TrackPlaytime !== false ? 'primary' : 'secondary'}"
			on:click={togglePlaytimeTracking}
		>
			<span class="material-icons mini-icon">{appSettings.TrackPlaytime !== false ? 'history' : 'history_toggle_off'}</span>
			<span>Playtime Tracking: {appSettings.TrackPlaytime !== false ? 'ON' : 'OFF'}</span>
		</button>
	</div>

	<div class="divider"></div>

	<div class="settings-section">
		<h3>Discord Rich Presence</h3>
		<p class="desc">
			Publish the active LightLauncher game and elapsed session time to Discord. The game instance owns the connection, so presence can remain active after the launcher window closes.
		</p>
		<button
			class="btn {appSettings.DiscordRichPresence !== false ? 'primary' : 'secondary'}"
			on:click={toggleDiscordPresence}
		>
			<span class="material-icons mini-icon">hub</span>
			<span>Discord Presence: {appSettings.DiscordRichPresence !== false ? 'ON' : 'OFF'}</span>
		</button>

		<label class="client-id-field">
			<span>Discord Application ID</span>
			<div class="client-id-row">
				<input
					class="input"
					type="text"
					bind:value={appSettings.DiscordClientId}
					placeholder="Discord application client ID"
					autocomplete="off"
					spellcheck="false"
				/>
				<button class="btn secondary" on:click={saveDiscordClientId} disabled={savingDiscordId}>
					<span class="material-icons mini-icon">save</span>
					{savingDiscordId ? 'Saving...' : 'Save'}
				</button>
			</div>
		</label>
		<p class="hint">You can also set <code>LIGHT_LAUNCHER_DISCORD_CLIENT_ID</code>. If Discord is closed or unavailable, game launch continues normally.</p>
	</div>
</div>

<style lang="scss">
	.zone-card {
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 28px;
	}

	.zone-header {
		display: flex;
		align-items: center;
		gap: 12px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);
		padding-bottom: 16px;

		.material-icons { color: var(--accent-primary); }
		h2 { margin: 0; font-size: 1.15rem; text-transform: uppercase; letter-spacing: 1px; }
	}

	.settings-section { display: flex; flex-direction: column; gap: 10px; }
	.settings-section h3 { margin: 0; font-size: 1.05rem; text-transform: uppercase; }
	.desc, .hint { margin: 0 0 8px; color: var(--text-muted); font-size: 0.88rem; line-height: 1.55; }
	.hint { margin-top: 2px; font-size: 0.78rem; color: var(--text-dim); }
	.divider { height: 2px; background: rgba(255, 255, 255, 0.05); border-radius: var(--radius-pill); }
	.mini-icon { font-size: 18px; margin-right: 6px; }

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
