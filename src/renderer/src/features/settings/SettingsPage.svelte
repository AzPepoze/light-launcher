<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { pendingSettingsSection } from "@stores/navigationStore";
	import { platform } from "@stores/platformStore";
	import { onMount } from "svelte";
	import { get } from "svelte/store";
	import type { PlatformInfo } from "@shared";
	import BlacklistSetting from "./BlacklistSetting.svelte";
	import GamingIntegrationsSetting from "./GamingIntegrationsSetting.svelte";
	import PrefixStorageSetting from "./PrefixStorageSetting.svelte";
	import ScanFoldersSetting from "./ScanFoldersSetting.svelte";
	import { SettingsState } from "./settingsState.svelte";

	const settings = new SettingsState();
	let platformInfo = $state<PlatformInfo>(get(platform));

	onMount(() => {
		void settings.start();
		return settings.subscribeUserSettings();
	});

	onMount(() => platform.subscribe((value) => (platformInfo = value)));

	onMount(() =>
		pendingSettingsSection.subscribe((section) => {
			if (settings.scrollToSection(section)) pendingSettingsSection.set(null);
		})
	);
</script>

<div class="settings-container" aria-busy={settings.loading}>
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
				<button class="btn" onclick={() => settings.toggleTheme()}>
					<span class="material-icons mini-icon">contrast</span>
					<span
						>Switch to {settings.userSettings.theme === "light"
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
						value={settings.userSettings.transparency}
						oninput={(event) => settings.setTransparency(parseFloat(event.currentTarget.value))}
						class="transparency-slider"
					/>
					<span class="pct-display"
						>{Math.round(settings.userSettings.transparency * 100)}%</span
					>
				</div>
				<div style="margin-top: 16px;">
					<button
						class="btn {settings.appSettings.TransparentMode ? 'primary' : 'secondary'}"
						onclick={() => settings.toggleTransparentMode()}
					>
						<span class="material-icons mini-icon"
							>{settings.appSettings.TransparentMode ? 'visibility' : 'visibility_off'}</span
						>
						<span
							>{settings.appSettings.TransparentMode
								? 'Transparent Background: ON'
								: 'Transparent Background: OFF'} (Restarts App)</span
						>
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

				{#if settings.userSettings.backgroundImagePath}
					<div class="bg-path-display">
						<span class="material-icons mini-icon">wallpaper</span>
						<span
							class="path-text"
							title={settings.userSettings.backgroundImagePath}
							>{settings.userSettings.backgroundImagePath}</span
						>
					</div>
				{/if}

				<div class="actions-row">
					<button class="btn primary" onclick={() => settings.browseBackgroundImage()}>
						<span class="material-icons mini-icon">folder</span>
						Browse Image
					</button>
					{#if settings.userSettings.backgroundImagePath}
						<button class="btn danger" onclick={() => settings.clearBackgroundImage()}>
							<span class="material-icons mini-icon">delete</span>
							Clear Image
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Zone 3: Display & Window Platform (Ozone, Linux only) -->
		{#if platformInfo.os === "linux"}
			<div class="zone-card">
				<div class="zone-header">
					<span class="material-icons">desktop_windows</span>
					<h2>Display & Window Platform</h2>
				</div>

				<div class="settings-section">
					<h3>Ozone Platform (Wayland / X11)</h3>
					<p class="desc">
						Configure whether LightLauncher runs natively on Wayland or via X11/Xwayland.
					</p>
					<div>
						<button
							class="btn {settings.appSettings.NativeWayland ? 'primary' : 'secondary'}"
							onclick={() => settings.toggleNativeWayland()}
						>
							<span class="material-icons mini-icon"
								>{settings.appSettings.NativeWayland ? 'layers' : 'view_in_ar'}</span
							>
							<span
								>{settings.appSettings.NativeWayland
									? 'Platform: Native Wayland'
									: 'Platform: X11 / Xwayland'} (Restarts App)</span
							>
						</button>
					</div>
				</div>
			</div>
		{/if}

		<GamingIntegrationsSetting {settings} />

		<!-- Wine Prefix Storage Location (Linux only) -->
		{#if platformInfo.features.prefix}
			<PrefixStorageSetting {settings} />
		{/if}

		<!-- Automatic Scan Folders -->
		<ScanFoldersSetting {settings} />

		<!-- Hidden / Blacklisted Games -->
		<BlacklistSetting {settings} />
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
