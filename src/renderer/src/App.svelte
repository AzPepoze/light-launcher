<script lang="ts">
	import {
		GetAppSettings,
		GetImageBase64,
		GetInitialGamePath,
		GetInitialLauncherPath,
		GetShouldEditLsfg,
	} from "@lib/api";
	import Navbar from "@components/shared/Navbar.svelte";
	import NotificationHost from "@components/shared/NotificationHost.svelte";
	import { navigationCommand } from "@stores/navigationStore";
	import { runState } from "@stores/runState";
	import { settingsStore } from "@stores/settingsStore";
	import { onMount } from "svelte";
	import { backOut } from "svelte/easing";
	import { fade, fly } from "svelte/transition";
	import EditLsfg from "./pages/EditLsfg.svelte";
	import Home from "./pages/Home.svelte";
	import Prefix from "./pages/Prefix.svelte";
	import Run from "./pages/Run.svelte";
	import Settings from "./pages/Settings.svelte";
	import Utils from "./pages/Utils.svelte";
	import Versions from "./pages/Versions.svelte";
	import CommandPalette from "@components/shared/CommandPalette.svelte";

	let bgBase64 = "";
	let transparency = 1.0;
	let theme: "light" | "dark" = "dark";

	$: bgColor = `rgba(var(--bg-base-rgb), ${transparency})`;

	settingsStore.subscribe(async (val) => {
		transparency = val.transparency;
		theme = val.theme;
		if (val.backgroundImagePath) {
			try {
				bgBase64 = await GetImageBase64(val.backgroundImagePath);
			} catch (e) {
				bgBase64 = "";
			}
		} else {
			bgBase64 = "";
		}
	});

	function toggleTheme() {
		settingsStore.update((s) => ({
			...s,
			theme: s.theme === "light" ? "dark" : "light",
		}));
	}

	let activePage = "home";
	let editLsfgGamePath = "";

	onMount(async () => {
		try {
			const appSettings = await GetAppSettings();
			document.documentElement.dataset.transparent =
				appSettings.TransparentMode.toString();

			const shouldEditLsfg = await GetShouldEditLsfg();
			const launcherPath = await GetInitialLauncherPath();

			if (shouldEditLsfg) {
				const gamePath = await GetInitialGamePath();
				if (gamePath) {
					editLsfgGamePath = gamePath;
					activePage = "editlsfg";
				}
			} else if (launcherPath) {
				runState.update((state) => ({
					...state,
					options: {
						...state.options,
						LauncherPath: launcherPath,
					},
				}));
				activePage = "run";
			}
		} catch (e) {
			console.error("Error in App onMount:", e);
		}
	});

	// Subscribe to navigation commands
	navigationCommand.subscribe((cmd) => {
		if (cmd) {
			if (cmd.page === "editlsfg" && cmd.gamePath) {
				editLsfgGamePath = cmd.gamePath;
				activePage = "editlsfg";
			} else if (cmd.page) {
				activePage = cmd.page;
			}
			navigationCommand.set(null);
		}
	});

	let showCommandPalette = false;

	function handleGlobalKeydown(e: KeyboardEvent) {
		if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
			showCommandPalette = !showCommandPalette;
			e.preventDefault();
		}
	}

	function handleNavigate(page: string) {
		activePage = page;
	}
</script>

<svelte:window on:keydown={handleGlobalKeydown} />

<main
	style="background-image: {bgBase64
		? `url(${bgBase64})`
		: 'none'}; background-size: cover; background-position: center; background-repeat: no-repeat; background-color: {bgColor};"
>
	<div class="app-layout" class:fullscreen={activePage === "editlsfg"}>
		{#if activePage !== "editlsfg"}
			<div class="navbar-container">
				<Navbar {activePage} onNavigate={handleNavigate} />
			</div>
		{/if}

		<div class="content-container">
			{#if activePage !== "editlsfg"}
				<div class="topbar-container">
					<button class="global-search-trigger" on:click={() => showCommandPalette = true} aria-label="Search games and actions">
						<span class="material-icons">search</span>
						<span class="placeholder-text">Search...</span>
						<span class="shortcut-kbd">Ctrl K</span>
					</button>
				</div>
			{/if}

			{#key activePage}
				<div
					class="page-wrapper"
					in:fly={{
						y: 30,
						duration: 400,
						delay: 100,
						easing: backOut,
					}}
					out:fade={{ duration: 150 }}
				>
					<div class="content-zone" class:full-width={activePage === "home" || activePage === "editlsfg"}>
						{#if activePage === "home"}
							<Home />
						{:else if activePage === "run"}
							<Run />
						{:else if activePage === "versions"}
							<Versions />
						{:else if activePage === "prefix"}
							<Prefix />
						{:else if activePage === "utils"}
							<Utils />
						{:else if activePage === "settings"}
							<Settings />
						{:else if activePage === "editlsfg"}
							<EditLsfg gamePath={editLsfgGamePath} />
						{:else}
							<div class="placeholder">
								Page "{activePage}" not implemented yet.
							</div>
						{/if}
					</div>
				</div>
			{/key}
		</div>
	</div>

	<NotificationHost />
	<CommandPalette bind:show={showCommandPalette} onClose={() => showCommandPalette = false} />
</main>

<style lang="scss">
	main {
		position: relative;
		height: 100vh;
		width: 100vw;
		color: var(--text-main);
		user-select: none;
		overflow: hidden;
	}

	.app-layout {
		display: flex;
		flex-direction: row !important;
		justify-content: flex-start !important;
		height: 100vh;
		width: 100vw;
		position: relative;
		z-index: 1;

		&.fullscreen {
			.navbar-container {
				display: none;
			}
		}
	}

	.navbar-container {
		position: fixed;
		left: 0;
		top: 0;
		width: 68px;
		height: 100vh;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
		align-items: stretch;
		z-index: 100;
		background: transparent;
		pointer-events: none;

		:global(*) {
			pointer-events: auto;
		}
	}

	.content-container {
		flex: 1;
		min-width: 0;
		height: 100%;
		position: relative;
		background: transparent;
	}

	.topbar-container {
		position: absolute;
		top: 24px;
		left: 50%;
		transform: translateX(-50%);
		width: 100%;
		max-width: 480px;
		display: flex;
		justify-content: center;
		z-index: 150;
		pointer-events: none;
	}

	.global-search-trigger {
		pointer-events: auto;
		display: inline-flex;
		align-items: center;
		width: 100%;
		gap: 8px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		padding: 8px 16px;
		color: var(--text-muted);
		cursor: pointer;
		font-size: 0.85rem;
		font-weight: 600;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		transition: border-color var(--transition-fast), background var(--transition-fast), color var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.15);
			background: var(--bg-elevated);
			color: var(--text-main);
		}

		.material-icons {
			font-size: 18px;
			color: var(--text-dim);
		}

		.placeholder-text {
			flex: 1;
			text-align: left;
		}

		.shortcut-kbd {
			font-size: 0.7rem;
			font-family: monospace;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			padding: 2px 6px;
			border-radius: var(--radius-sm);
			color: var(--text-dim);
			margin-left: 12px;
		}
	}

	.page-wrapper {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		overflow-y: auto;
		padding: 76px 48px 40px 84px;
		box-sizing: border-box;
	}

	.content-zone {
		width: 100%;
		max-width: var(--content-max-width, 1150px);
		margin: 0 auto;
		box-sizing: border-box;

		&.full-width {
			max-width: 100%;
		}
	}

	.placeholder {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--text-dim);
		font-size: 0.9rem;
		font-style: italic;
	}
</style>
