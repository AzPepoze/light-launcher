<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import RangeSlider from "@components/shared/RangeSlider.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import Slider from "@components/shared/Slider.svelte";
	import { tooltip } from "@components/shared/tooltip";
	import { activeSectionId, navSectionIds, SPY_LINE_PX } from "@lib/features/settingsScrollSpy";
	import {
		settingsNavFilter,
		settingsNavGroups,
		type SettingsNavGroup
	} from "@lib/features/settingsSection";
	import { pendingSettingsSection } from "@stores/navigationStore";
	import { platform } from "@stores/platformStore";
	import { onMount } from "svelte";
	import { get } from "svelte/store";
	import type { PlatformInfo } from "@shared";
	import BlacklistSetting from "./BlacklistSetting.svelte";
	import DownloadsSettings from "./DownloadsSettings.svelte";
	import GamingIntegrationsSetting from "./GamingIntegrationsSetting.svelte";
	import PluginRegistrySetting from "./PluginRegistrySetting.svelte";
	import PrefixStorageSetting from "./PrefixStorageSetting.svelte";
	import ScanFoldersSetting from "./ScanFoldersSetting.svelte";
	import SettingsNav from "./SettingsNav.svelte";
	import TrustedKeysSetting from "./TrustedKeysSetting.svelte";
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
	let scroller = $state<HTMLElement>();
	let activeId = $state<string | null>(null);

	const navGroups = $derived(
		settingsNavGroups({
			linux: platformInfo.os === "linux",
			prefix: platformInfo.features.prefix
		})
	);
	const navIds = $derived(navSectionIds(navGroups));

	const SEARCH_PLACEHOLDER = "Search settings";

	interface SearchResult {
		id: string;
		label: string;
		group: string;
	}

	let query = $state("");
	const searching = $derived(query.trim() !== "");
	const results = $derived(searching ? flattenMatches(settingsNavFilter(navGroups, query)) : []);
	const hasNoResults = $derived(searching && results.length === 0);

	function flattenMatches(groups: SettingsNavGroup[]): SearchResult[] {
		return groups.flatMap((group) =>
			group.items.map((item) => ({ id: item.id, label: item.label, group: group.label }))
		);
	}

	function jumpFromSearch(id: string): void {
		jumpTo(id);
		query = "";
	}

	function onSearchKeydown(event: KeyboardEvent): void {
		if (event.key === "Enter" && results[0]) {
			event.preventDefault();
			jumpFromSearch(results[0].id);
		} else if (event.key === "Escape") {
			query = "";
		}
	}

	function toPercent(value: number): string {
		return `${Math.round(value * 100)}%`;
	}

	function measureTops(root: HTMLElement, ids: string[]): Record<string, number> {
		const origin = root.getBoundingClientRect().top;
		const tops: Record<string, number> = {};
		for (const id of ids) {
			const element = document.getElementById(id);
			if (element) tops[id] = element.getBoundingClientRect().top - origin;
		}
		return tops;
	}

	function isAtBottom(root: HTMLElement): boolean {
		return root.scrollTop + root.clientHeight >= root.scrollHeight - 4;
	}

	function refreshActive(): void {
		if (!scroller) return;
		const tops = measureTops(scroller, navIds);
		activeId = activeSectionId(navIds, tops, SPY_LINE_PX, isAtBottom(scroller));
	}

	function jumpTo(id: string): void {
		activeId = id;
		settings.scrollToSection(id);
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

	// The observer covers sections entering the top of the scroll area. The scroll
	// handler covers the bottom of the page, where a short last section never crosses the line.
	$effect(() => {
		const root = scroller;
		if (!root) return;
		const observer = new IntersectionObserver(refreshActive, {
			root,
			rootMargin: "-24px 0px -60% 0px"
		});
		for (const id of navIds) {
			const element = document.getElementById(id);
			if (element) observer.observe(element);
		}
		refreshActive();
		return () => observer.disconnect();
	});
</script>

<div class="settings-page" aria-busy={settings.loading}>
	<PageHeader title="Appearance & Settings" icon="settings" />

	<div class="layout">
		<div class="nav-column">
			<SettingsNav groups={navGroups} active={activeId} onSelect={jumpTo} />
		</div>

		<div class="sections-column">
		<div class="search-bar">
			<input
				class="input search"
				type="search"
				aria-label={SEARCH_PLACEHOLDER}
				placeholder={SEARCH_PLACEHOLDER}
				bind:value={query}
				onkeydown={onSearchKeydown}
			/>

			{#if searching}
				<div class="results">
					{#if hasNoResults}
						<p class="no-match">No matching sections.</p>
					{:else}
						<ul>
							{#each results as result (result.id)}
								<li>
									<button
										type="button"
										class="result"
										onclick={() => jumpFromSearch(result.id)}
									>
										<span class="result-label">{result.label}</span>
										<span class="result-group">{result.group}</span>
									</button>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
		</div>

		<div class="zones-stack" bind:this={scroller} onscroll={refreshActive}>
			<div class="anchor" id="theme-opacity">
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

					<SlideButton
						label="Transparent background"
						subtitle="Restarts the app to apply"
						checked={settings.appSettings.TransparentMode}
						onChange={(enabled) => settings.setTransparentMode(enabled)}
					/>
				</Section>
			</div>

			<div class="anchor" id="wallpaper">
				<Section
					title="Wallpaper background"
					icon="wallpaper"
					help="Set a custom image background for the launcher that bleeds through the translucent glass surfaces."
				>
					{#if settings.userSettings.backgroundImagePath}
						<div class="bg-path-display">
							<Icon name="wallpaper" size={18} />
							<span class="path-text" use:tooltip={settings.userSettings.backgroundImagePath}>
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
			</div>

			<div class="anchor" id="library-covers">
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
			</div>

			<div class="anchor" id="game-folders">
				<ScanFoldersSetting {settings} />
			</div>

			<div class="anchor" id="hidden-games">
				<BlacklistSetting {settings} />
			</div>

			<div class="anchor" id="game-activity">
				<GamingIntegrationsSetting {settings} />
			</div>

			<div class="anchor" id="downloads">
				<DownloadsSettings />
			</div>

			<div class="anchor" id="updates">
				<UpdatesSettings {settings} />
			</div>

			<div class="anchor" id="trusted-keys">
				<TrustedKeysSetting />
			</div>

			<div class="anchor" id="community-plugins">
				<PluginRegistrySetting {settings} />
			</div>

			{#if platformInfo.features.prefix}
				<div class="anchor" id="prefix-storage">
					<PrefixStorageSetting {settings} />
				</div>
			{/if}

			<div class="anchor" id="window-type">
				<Section title="Window type" icon="open_in_browser">
					<SlideButton
						label="Open in browser"
						subtitle="Use your default browser instead of the app window. Applies on the next start."
						checked={settings.appSettings.OpenInBrowser}
						onChange={(enabled) => settings.setOpenInBrowser(enabled)}
					/>
				</Section>
			</div>

			{#if platformInfo.os === "linux"}
				<div class="anchor" id="display-platform">
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
				</div>
			{/if}
		</div>
		</div>
	</div>
</div>

<style lang="scss">
	.settings-page {
		height: 100%;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		padding: 40px 48px 0;
	}

	/* Fills the page height so the nav stays put and only the sections column scrolls. */
	.layout {
		flex: 1;
		min-height: 0;
		display: flex;
		align-items: stretch;
		gap: 32px;
		margin-top: 24px;
	}

	.nav-column {
		flex: 0 0 200px;
		min-height: 0;
		overflow-y: auto;
	}

	/* The search bar stays above the scroll area, so the sections scroll under it. */
	.sections-column {
		flex: 1;
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.search-bar {
		position: relative;
		flex: none;
	}

	.search {
		width: 100%;
		box-sizing: border-box;
		padding: 10px 16px;
		font-size: 0.9rem;
	}

	/* Floats over the sections, so it does not push them down. */
	.results {
		position: absolute;
		top: calc(100% + 8px);
		left: 0;
		right: 0;
		z-index: 20;
		max-height: 50vh;
		overflow-y: auto;
		padding: 8px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
	}

	.results ul {
		display: flex;
		flex-direction: column;
		gap: 2px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.result {
		display: flex;
		align-items: baseline;
		flex-wrap: wrap;
		gap: 8px;
		width: 100%;
		padding: 10px 12px;
		border: 0;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-main);
		font: inherit;
		font-size: 0.9rem;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
		transition: background var(--transition-fast);

		&:hover,
		&:focus-visible {
			background: var(--bg-elevated);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.result-group {
		color: var(--text-dim);
		font-size: 0.8rem;
		font-weight: 600;

		&::before {
			content: "·";
			margin-right: 8px;
		}
	}

	.no-match {
		margin: 0;
		padding: 10px 12px;
		color: var(--text-dim);
		font-size: 0.85rem;
	}

	.zones-stack {
		flex: 1;
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		gap: 32px;
		padding-bottom: 40px;
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

	/* Keeps a section clear of the top edge when the nav jumps to it. */
	.anchor {
		scroll-margin-top: 24px;
	}

	/* Under 900px the nav becomes a chip row above the sections, which keep their own scroll. */
	@media (max-width: 899px) {
		.layout {
			flex-direction: column;
			gap: 24px;
		}

		.nav-column {
			flex: none;
			width: 100%;
			overflow: visible;
		}
	}

	@media (max-width: 599px) {
		.settings-page {
			padding: 24px 16px 0;
		}

		/* Keeps each section's header and rows inside the phone width. */
		.zones-stack :global(.section) {
			padding: 20px 16px;
		}

		.zones-stack :global(.section-header) {
			flex-wrap: wrap;
		}

		.zones-stack :global(.setting-row) {
			flex-direction: column;
			align-items: stretch;
			gap: 12px;
		}

		.zones-stack :global(.setting-row .control) {
			flex-wrap: wrap;
			min-width: 0;
		}

		.zones-stack :global(.setting-row .button) {
			max-width: 100%;
			white-space: normal;
			text-align: center;
		}

		.zones-stack :global(.title),
		.zones-stack :global(.help),
		.zones-stack :global(.slide-button-card .title),
		.zones-stack :global(.slide-button-card .subtitle) {
			overflow-wrap: anywhere;
		}
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
			min-width: 0;
			font-size: 0.85rem;
			font-weight: 600;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	.actions-row {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
	}
</style>
