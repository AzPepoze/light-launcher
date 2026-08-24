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
	import { cubicOut } from "svelte/easing";
	import { fly } from "svelte/transition";
	import EditLsfg from "./pages/EditLsfg.svelte";
	import Home from "./pages/Home.svelte";
	import GameManager from "./pages/GameManager.svelte";
	import Prefix from "./pages/Prefix.svelte";
	import Run from "./pages/Run.svelte";
	import Settings from "./pages/Settings.svelte";
	import Utils from "./pages/Utils.svelte";
	import Versions from "./pages/Versions.svelte";
	import CommandPalette from "@components/shared/CommandPalette.svelte";
	import { commandPaletteState } from "@components/shared/CommandPaletteState.svelte";

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

	const PAGE_ORDER: Record<string, number> = {
		home: 0,
		manager: 1,
		run: 2,
		versions: 3,
		prefix: 4,
		utils: 5,
		settings: 6,
		editlsfg: 7,
	};

	let activePage = "home";
	let navDirection = 1;
	let editLsfgGamePath = "";

	function changePage(newPage: string) {
		if (!newPage || newPage === activePage) return;
		const prevIndex = PAGE_ORDER[activePage] ?? 0;
		const newIndex = PAGE_ORDER[newPage] ?? 0;
		navDirection = newIndex >= prevIndex ? 1 : -1;
		activePage = newPage;
		scrolled = false;
	}

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
					changePage("editlsfg");
				}
			} else if (launcherPath) {
				runState.update((state) => ({
					...state,
					options: {
						...state.options,
						LauncherPath: launcherPath,
					},
				}));
				changePage("run");
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
				changePage("editlsfg");
			} else if (cmd.page) {
				changePage(cmd.page);
			}
			navigationCommand.set(null);
		}
	});

	let showCommandPalette = false;

	function openCommandPalette() {
		showCommandPalette = true;
		setTimeout(() => commandPaletteState.focusInput(), 0);
	}

	function handleGlobalKeydown(e: KeyboardEvent) {
		if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
			openCommandPalette();
			e.preventDefault();
		}
	}

	function handleNavigate(page: string) {
		changePage(page);
	}

	let scrolled = false;

	function handlePageScroll(e: Event) {
		scrolled = (e.currentTarget as HTMLElement).scrollTop > 0;
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
					<div class="global-search-trigger" class:open={showCommandPalette} class:scrolled={scrolled} role="search">
						<span class="material-icons">search</span>
						<input
							bind:this={commandPaletteState.inputElement}
							bind:value={commandPaletteState.searchQuery}
							type="text"
							placeholder="Search games, pages, and actions..."
							aria-label="Search games and actions"
							autocomplete="off"
							spellcheck="false"
							on:focus={openCommandPalette}
							on:input={openCommandPalette}
						/>
						<button class="shortcut-kbd" on:click={openCommandPalette} aria-label="Open command palette">Ctrl K</button>
					</div>
					<CommandPalette bind:show={showCommandPalette} onClose={() => showCommandPalette = false} />
				</div>
			{/if}

			{#key activePage}
				<div
					class="page-wrapper"
					class:home-mode={activePage === "home"}
					on:scroll={handlePageScroll}
					in:fly={{
						y: 35 * navDirection,
						duration: 280,
						easing: cubicOut,
					}}
					out:fly={{
						y: -35 * navDirection,
						duration: 200,
						easing: cubicOut,
					}}
				>
					<div class="content-zone" class:full-width={activePage === "home" || activePage === "editlsfg" || activePage === "manager"} class:home-zone={activePage === "home"}>
						{#if activePage === "home"}
							<Home />
						{:else if activePage === "manager"}
							<GameManager />
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
		overflow: hidden;
	}

	.topbar-container {
		position: absolute;
		top: 24px;
		left: 50%;
		transform: translateX(-50%);
		width: 100%;
		max-width: 520px;
		display: flex;
		flex-direction: column;
		align-items: center;
		z-index: 150;
		pointer-events: none;
	}

	.global-search-trigger {
		pointer-events: auto;
		display: flex;
		align-items: center;
		width: 100%;
		gap: 8px;
		border: 2px solid transparent;
		border-radius: var(--radius-md);
		padding: 6px 10px 6px 14px;
		color: var(--text-muted);
		transition: border-color var(--transition-fast), background var(--transition-fast), box-shadow var(--transition-fast);

		&.scrolled, &.open {
			background: var(--bg-surface);
			border-color: rgba(255, 255, 255, 0.05);
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		}

		&.open {
			border-color: rgba(255, 255, 255, 0.16);
			border-bottom-left-radius: 0;
			border-bottom-right-radius: 0;
			box-shadow: none;
		}

		&:focus-within, &:hover {
			border-color: rgba(255, 255, 255, 0.16);
			background: var(--bg-elevated);
			box-shadow: 0 6px 24px rgba(0, 0, 0, 0.22);
		}

		.material-icons {
			font-size: 18px;
			color: var(--text-dim);
		}

		input {
			flex: 1;
			min-width: 0;
			border: 0;
			outline: 0;
			background: transparent;
			color: var(--text-main);
			font: inherit;
			font-size: 0.88rem;
			font-weight: 650;
			user-select: text;

			&::placeholder { color: var(--text-dim); }
		}

		.shortcut-kbd {
			font-size: 0.7rem;
			font-family: monospace;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			padding: 4px 7px;
			border-radius: var(--radius-sm);
			color: var(--text-dim);
			cursor: pointer;
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

		&.home-mode {
			overflow: hidden;
			display: flex;
			flex-direction: column;
			padding-bottom: 0;
		}
	}

	.content-zone {
		width: 100%;
		max-width: var(--content-max-width, 1150px);
		margin: 0 auto;
		box-sizing: border-box;

		&.full-width {
			max-width: 100%;
		}

		&.home-zone {
			flex: 1;
			min-height: 0;
			display: flex;
			flex-direction: column;
			overflow: hidden;
			height: 100%;
			max-height: 100%;
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
