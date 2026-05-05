<script lang="ts">
	import Navbar from "@components/shared/Navbar.svelte";
	import Home from "./pages/Home.svelte";
	import Run from "./pages/Run.svelte";
	import Versions from "./pages/Versions.svelte";
	import Prefix from "./pages/Prefix.svelte";
	import Utils from "./pages/Utils.svelte";
	import Settings from "./pages/Settings.svelte";
	import EditLsfg from "./pages/EditLsfg.svelte";
	import NotificationHost from "@components/shared/NotificationHost.svelte";
	import { fade, fly } from "svelte/transition";
	import {
		GetInitialLauncherPath,
		GetInitialGamePath,
		GetShouldEditLsfg,
		GetImageBase64,
		GetAppSettings,
	} from "@bindings/light-launcher/internal/app/app";
	import { onMount } from "svelte";
	import { navigationCommand } from "@stores/navigationStore";
	import { runState } from "@stores/runState";
	import { settingsStore } from "@stores/settingsStore";

	let bgBase64 = "";
	let transparency = 1.0;

	settingsStore.subscribe(async (val) => {
		transparency = val.transparency;
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
			document.documentElement.dataset.transparent = appSettings.TransparentMode.toString();

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

	function handleNavigate(page: string) {
		activePage = page;
	}
</script>

<main style="background-image: {bgBase64 ? `url(${bgBase64})` : 'none'}; background-size: cover; background-position: center; background-repeat: no-repeat;">
	<div class="app-layout" class:fullscreen={activePage === "editlsfg"}>
		{#if activePage !== "editlsfg"}
			<div class="navbar-container">
				<Navbar
					{activePage}
					onNavigate={handleNavigate}
					{toggleTheme}
				/>
			</div>
		{/if}

		<div class="content-container">
			{#key activePage}
				<div
					class="page-wrapper"
					in:fly={{ y: 10, duration: 300, delay: 150 }}
					out:fade={{ duration: 150 }}
				>
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
			{/key}
		</div>
	</div>

	<NotificationHost />
</main>

<style lang="scss">
	main {
		position: relative;
		height: 100vh;
		width: 100vw;
		background-color: var(--glass-bg);
		color: var(--text-main);
		user-select: none;
		overflow: hidden; /* Prevent scrollbar flicker during transition */
	}

	.app-layout {
		display: flex;
		flex-direction: row !important; /* Force horizontal layout */
		justify-content: flex-start !important; /* Start from left */
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
		order: 0 !important; /* Absolute first position */
		width: 80px; /* Thinner sidebar area */
		padding: 24px 0;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		z-index: 10;
		background: transparent;
	}

	.content-container {
		order: 1 !important; /* Absolute second position */
		flex: 1;
		min-width: 0;
		height: 100%;
		position: relative;
		background: transparent;
	}

	/* Wrapper to handle transition positioning */
	.page-wrapper {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		overflow-y: auto; /* Allow scrolling inside the page */
		padding: 24px;
		box-sizing: border-box;
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
