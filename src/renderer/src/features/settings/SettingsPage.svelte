<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import RangeSlider from "@components/shared/RangeSlider.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import Slider from "@components/shared/Slider.svelte";
	import { pendingSettingsSection } from "@stores/navigationStore";
	import { platform } from "@stores/platformStore";
	import { onMount } from "svelte";
	import { get } from "svelte/store";
	import type { PlatformInfo } from "@shared";
	import BlacklistSetting from "./BlacklistSetting.svelte";
	import DownloadsSettings from "./DownloadsSettings.svelte";
	import GamingIntegrationsSetting from "./GamingIntegrationsSetting.svelte";
	import PrefixStorageSetting from "./PrefixStorageSetting.svelte";
	import ScanFoldersSetting from "./ScanFoldersSetting.svelte";
	import UpdatesSettings from "./UpdatesSettings.svelte";
	import {
		LIBRARY_CARD_SIZE_MAX,
		LIBRARY_CARD_SIZE_MIN,
		blurCoversUntilHover,
		libraryCardSize
	} from "./libraryAppearance";
	import { SettingsState } from "./settingsState.svelte";

	const settings = new SettingsState();
	let platformInfo = $state<PlatformInfo>(get(platform));

	function toPercent(value: number): string {
		return `${Math.round(value * 100)}%`;
	}

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
		<Section title="Theme & opacity" icon="palette">
			<SettingRow title="Theme mode" help="Switch between dark and light themes.">
				<Button variant="secondary" icon="contrast" onclick={() => settings.toggleTheme()}>
					Switch to {settings.userSettings.theme === "light" ? "Dark" : "Light"} Mode
				</Button>
			</SettingRow>

			<SettingRow
				title="Background opacity"
				help="Lower values make the background more see-through while text and UI stay fully visible."
			>
				<Slider
					value={settings.userSettings.transparency}
					min={0.1}
					max={1}
					step={0.05}
					format={toPercent}
					onChange={(value) => settings.setTransparency(value)}
				/>
			</SettingRow>

			<SettingRow title="Transparent background" help="Restarts the app to apply.">
				<Button
					variant={settings.appSettings.TransparentMode ? "primary" : "secondary"}
					icon={settings.appSettings.TransparentMode ? "visibility" : "visibility_off"}
					onclick={() => settings.toggleTransparentMode()}
				>
					{settings.appSettings.TransparentMode
						? "Transparent Background: ON"
						: "Transparent Background: OFF"} (Restarts App)
				</Button>
			</SettingRow>
		</Section>

		<Section
			title="Wallpaper background"
			icon="wallpaper"
			help="Set a custom image background for the launcher that bleeds through the translucent glass surfaces."
		>
			{#if settings.userSettings.backgroundImagePath}
				<div class="bg-path-display">
					<Icon name="wallpaper" size={18} />
					<span class="path-text" title={settings.userSettings.backgroundImagePath}>
						{settings.userSettings.backgroundImagePath}
					</span>
				</div>
			{/if}

			<div class="actions-row">
				<Button variant="primary" icon="folder" onclick={() => settings.browseBackgroundImage()}>
					Browse image
				</Button>
				{#if settings.userSettings.backgroundImagePath}
					<Button variant="danger" icon="delete" onclick={() => settings.clearBackgroundImage()}>
						Clear image
					</Button>
				{/if}
			</div>
		</Section>

		<Section title="Library covers" icon="grid_view">
			<SettingRow title="Card size" help="Minimum width of a Library card in pixels.">
				<RangeSlider
					value={libraryCardSize(settings.appSettings)}
					min={LIBRARY_CARD_SIZE_MIN}
					max={LIBRARY_CARD_SIZE_MAX}
					unit="px"
					label={`${libraryCardSize(settings.appSettings)} px`}
					snapValues={[]}
					showWarning={false}
					onChange={(value) => settings.setLibraryCardSize(value)}
				/>
			</SettingRow>

			<SlideButton
				label="Blur covers until hover"
				subtitle="Blur Library covers until you hover or focus them."
				checked={blurCoversUntilHover(settings.appSettings)}
				onChange={(enabled) => settings.toggleBlurCoversUntilHover(enabled)}
			/>
		</Section>

		<DownloadsSettings />

		<UpdatesSettings {settings} />

		{#if platformInfo.os === "linux"}
			<Section title="Display & window platform" icon="desktop_windows">
				<SettingRow
					title="Ozone platform (Wayland / X11)"
					help="Configure whether LightLauncher runs natively on Wayland or via X11/Xwayland."
				>
					<Button
						variant={settings.appSettings.NativeWayland ? "primary" : "secondary"}
						icon={settings.appSettings.NativeWayland ? "layers" : "view_in_ar"}
						onclick={() => settings.toggleNativeWayland()}
					>
						{settings.appSettings.NativeWayland
							? "Platform: Native Wayland"
							: "Platform: X11 / Xwayland"} (Restarts App)
					</Button>
				</SettingRow>
			</Section>
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

	.bg-path-display {
		display: flex;
		align-items: center;
		gap: 6px;
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
	}
</style>
