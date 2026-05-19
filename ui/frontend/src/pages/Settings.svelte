<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { settingsStore } from "@stores/settingsStore";
	import * as service from "@lib/settingsService";
	import { onMount } from "svelte";
	import ScanFoldersSetting from "@components/settings/ScanFoldersSetting.svelte";
	import BlacklistSetting from "@components/settings/BlacklistSetting.svelte";

	let currentSettings = {
		theme: "light",
		transparency: 1.0,
		backgroundImagePath: "",
	};

	let appSettings = {
		TransparentMode: true,
		ScanFolders: [] as string[],
		Blacklist: [] as string[],
	};

	async function refreshAppSettings() {
		const settings = await service.loadAppSettings();
		if (settings) {
			appSettings = settings;
		}
	}

	onMount(async () => {
		await refreshAppSettings();
	});

	settingsStore.subscribe((val) => {
		currentSettings = val;
	});

	function handleThemeToggle() {
		service.toggleTheme();
	}

	function handleTransparencyChange(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		service.updateTransparency(val);
	}



	async function handleBrowseBackground() {
		await service.browseBackgroundImage();
	}

	function handleClearBackground() {
		service.clearBackgroundImage();
	}

	async function toggleTransparentMode() {
		await service.toggleTransparentMode(appSettings);
	}
</script>

<div class="settings-container">
	<PageHeader title="Appearance & Settings" icon="settings" />

	<div class="zones-stack">
		<!-- Zone 1: Theme & Opacity -->
		<div class="zone-card">
			<div class="zone-header">
				<span class="material-icons">palette</span>
				<h2>Theme & Opacity</h2>
			</div>

			<div class="settings-section">
				<h3>Theme Mode</h3>
				<p class="desc">Switch between dark and light themes.</p>
				<button class="btn" on:click={handleThemeToggle}>
					<span class="material-icons mini-icon">contrast</span>
					<span
						>Switch to {currentSettings.theme === "light"
							? "Dark"
							: "Light"} Mode</span
					>
				</button>
			</div>

			<div class="divider"></div>

			<div class="settings-section">
				<h3>Background Opacity</h3>
				<p class="desc">
					Adjust the background opacity. Lower values make the background more see-through while keeping text and UI elements fully visible.
				</p>
				<div class="slider-row">
					<input
						type="range"
						min="0.1"
						max="1.0"
						step="0.05"
						value={currentSettings.transparency}
						on:input={handleTransparencyChange}
						class="transparency-slider"
					/>
					<span class="pct-display"
						>{Math.round(currentSettings.transparency * 100)}%</span
					>
				</div>
				<div style="margin-top: 16px;">
					<button class="btn {appSettings.TransparentMode ? 'primary' : 'secondary'}" on:click={toggleTransparentMode}>
						<span class="material-icons mini-icon">{appSettings.TransparentMode ? 'visibility' : 'visibility_off'}</span>
						<span>{appSettings.TransparentMode ? 'Transparent Background: ON' : 'Transparent Background: OFF'} (Restarts App)</span>
					</button>
				</div>
			</div>
		</div>

		<!-- Zone 2: Wallpaper Background -->
		<div class="zone-card">
			<div class="zone-header">
				<span class="material-icons">wallpaper</span>
				<h2>Wallpaper Background</h2>
			</div>

			<div class="settings-section">
				<p class="desc">
					Set a custom image background for the launcher that bleeds through the translucent glass surfaces.
				</p>

				{#if currentSettings.backgroundImagePath}
					<div class="bg-path-display">
						<span class="material-icons mini-icon">wallpaper</span>
						<span
							class="path-text"
							title={currentSettings.backgroundImagePath}
							>{currentSettings.backgroundImagePath}</span
						>
					</div>
				{/if}

				<div class="actions-row">
					<button class="btn primary" on:click={handleBrowseBackground}>
						<span class="material-icons mini-icon">folder</span>
						Browse Image
					</button>
					{#if currentSettings.backgroundImagePath}
						<button class="btn danger" on:click={handleClearBackground}>
							<span class="material-icons mini-icon">delete</span>
							Clear Image
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Zone 3: Automatic Scan Folders -->
		<ScanFoldersSetting {appSettings} onRefresh={refreshAppSettings} />

		<!-- Zone 4: Hidden / Blacklisted Games -->
		<BlacklistSetting {appSettings} onRefresh={refreshAppSettings} />
	</div>
</div>

<style lang="scss">
	.settings-container {
		padding: 40px 48px;
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow-y: auto;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 102, 171, 0.2);
			border-radius: 10px;
		}
	}

	.zones-stack {
		display: flex;
		flex-direction: column;
		gap: 32px;
		margin-top: 24px;
	}

	.zone-card {
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 28px;
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
		}

		.zone-header {
			display: flex;
			align-items: center;
			gap: 12px;
			border-bottom: 2px solid rgba(255, 255, 255, 0.05);
			padding-bottom: 16px;
			margin-bottom: 4px;

			.material-icons {
				font-size: 24px;
				color: var(--accent-primary);
			}

			h2 {
				margin: 0;
				font-size: 1.15rem;
				font-weight: 800;
				color: var(--text-main);
				text-transform: uppercase;
				letter-spacing: 1px;
			}
		}
	}

	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 8px;

		h3 {
			margin: 0;
			font-size: 1.2rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.desc {
			margin: 0 0 12px 0;
			font-size: 0.9rem;
			font-weight: 600;
			color: var(--text-muted);
		}

		.mini-icon {
			font-size: 18px;
			margin-right: 6px;
		}
	}

	.slider-row {
		display: flex;
		align-items: center;
		gap: 16px;

		.transparency-slider {
			flex: 1;
			accent-color: var(--accent-primary);
			height: 6px;
			border-radius: var(--radius-pill);
			background: var(--bg-input);
			outline: none;
		}

		.pct-display {
			font-size: 1rem;
			font-weight: 800;
			color: var(--accent-primary);
			width: 45px;
			text-align: right;
		}
	}

	.bg-path-display {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		background: var(--bg-base);
		border: 2px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-md);
		margin-bottom: 16px;
		color: var(--text-muted);

		.path-text {
			font-size: 0.85rem;
			font-weight: 600;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	.actions-row {
		display: flex;
		gap: 12px;

		button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}
	}

	.divider {
		height: 2px;
		background: rgba(255, 255, 255, 0.05);
		margin: 12px 0;
		border-radius: var(--radius-pill);
	}
</style>
