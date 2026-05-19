<script lang="ts">
	import { GetAllGames, RunGame } from "@bindings/light-launcher/internal/app/app";
	import { navigationCommand } from "@stores/navigationStore";
	import { notifications } from "@stores/notificationStore";
	import { fade } from "svelte/transition";
	import { loadExeIcon } from "@lib/iconService";
	import protonIcon from "@icons/protron_forked.png";

	export let show = false;
	export let onClose: () => void = () => {};

	let searchQuery = "";
	let games: any[] = [];
	let gameIcons: Record<string, string> = {};
	let filteredItems: any[] = [];
	let selectedIndex = 0;
	let inputElement: HTMLInputElement;

	const PAGES = [
		{ name: "Go to Home", icon: "home", action: () => navigateTo("home") },
		{ name: "Go to Launch Configuration", icon: "play_arrow", action: () => navigateTo("run") },
		{ name: "Go to Utilities", icon: "handyman", action: () => navigateTo("utils") },
		{ name: "Go to Appearance & Settings", icon: "settings", action: () => navigateTo("settings") },
		{ name: "Go to Proton Versions", icon: protonIcon, isCustomIcon: true, action: () => navigateTo("versions") },
		{ name: "Go to WINE Prefixes", icon: "folder", action: () => navigateTo("prefix") }
	];

	async function loadGames() {
		try {
			const fetched = await GetAllGames();
			games = fetched || [];
			
			for (const game of games) {
				const path = game.path || game.config?.LauncherPath;
				if (path && !gameIcons[path]) {
					loadExeIcon(path).then((icon) => {
						if (icon) {
							gameIcons[path] = icon;
							gameIcons = gameIcons; // trigger reactivity
						}
					});
				}
			}

			filterItems();
		} catch (e) {
			console.error("Failed to load games for command palette", e);
		}
	}

	function filterItems() {
		const query = searchQuery.trim().toLowerCase();
		
		// If empty, show pages first, then some games
		if (!query) {
			filteredItems = [
				...PAGES.map(p => ({ ...p, type: "page" })),
				...games.slice(0, 5).map(g => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }))
			];
			selectedIndex = 0;
			return;
		}

		// Otherwise filter
		const matchedPages = PAGES.filter(p => p.name.toLowerCase().includes(query)).map(p => ({ ...p, type: "page" }));
		const matchedGames = games.filter(g => g.name.toLowerCase().includes(query)).map(g => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }));
		
		filteredItems = [...matchedPages, ...matchedGames];
		selectedIndex = 0;
	}

	function navigateTo(page: string) {
		navigationCommand.set({ page });
		close();
	}

	async function executeItem(item: any) {
		if (item.type === "page") {
			item.action();
		} else if (item.type === "game") {
			try {
				notifications.add(`Launching ${item.game.name}...`, "info");
				close();
				await RunGame(item.game.config, false);
			} catch (err) {
				notifications.add(`Launch failed: ${err}`, "error");
			}
		}
	}

	function close() {
		onClose();
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape") {
			close();
			e.preventDefault();
		} else if (e.key === "ArrowDown") {
			selectedIndex = (selectedIndex + 1) % Math.max(1, filteredItems.length);
			e.preventDefault();
			scrollToSelected();
		} else if (e.key === "ArrowUp") {
			selectedIndex = (selectedIndex - 1 + filteredItems.length) % Math.max(1, filteredItems.length);
			e.preventDefault();
			scrollToSelected();
		} else if (e.key === "Enter") {
			if (filteredItems[selectedIndex]) {
				executeItem(filteredItems[selectedIndex]);
			}
			e.preventDefault();
		}
	}

	let resultsContainer: HTMLDivElement;
	function scrollToSelected() {
		if (!resultsContainer) return;
		const selectedEl = resultsContainer.children[selectedIndex] as HTMLElement;
		if (!selectedEl) return;

		const containerHeight = resultsContainer.clientHeight;
		const elTop = selectedEl.offsetTop;
		const elHeight = selectedEl.clientHeight;

		if (elTop < resultsContainer.scrollTop) {
			resultsContainer.scrollTop = elTop;
		} else if (elTop + elHeight > resultsContainer.scrollTop + containerHeight) {
			resultsContainer.scrollTop = elTop + elHeight - containerHeight;
		}
	}

	$: if (show) {
		searchQuery = "";
		selectedIndex = 0;
		loadGames();
		setTimeout(() => {
			if (inputElement) inputElement.focus();
		}, 50);
	}

	$: if (searchQuery !== undefined) {
		filterItems();
	}
</script>

{#if show}
	<button class="palette-backdrop" on:click={close} on:keydown={(e) => e.key === "Escape" && close()} transition:fade={{ duration: 150 }} aria-label="Close palette">
		<div class="palette-container" on:click|stopPropagation on:keydown|stopPropagation={() => {}} role="dialog" aria-modal="true" tabindex="-1">
			<div class="search-row">
				<span class="material-icons search-icon">search</span>
				<input
					bind:this={inputElement}
					bind:value={searchQuery}
					on:keydown={handleKeyDown}
					placeholder="Search games, pages, and actions..."
					type="text"
				/>
				<span class="esc-badge">ESC</span>
			</div>

			<div class="results-list" bind:this={resultsContainer}>
				{#each filteredItems as item, i}
					<div
						class="result-item"
						class:active={i === selectedIndex}
						on:click={() => executeItem(item)}
						on:mouseenter={() => selectedIndex = i}
						role="button"
						tabindex="0"
						on:keydown={(e) => e.key === "Enter" && executeItem(item)}
					>
						{#if item.type === "game" && gameIcons[item.game.path || item.game.config?.LauncherPath]}
							<img src={gameIcons[item.game.path || item.game.config?.LauncherPath]} class="item-img-icon" alt="" />
						{:else if item.isCustomIcon}
							<img src={item.icon} class="item-img-icon" alt="" />
						{:else}
							<span class="material-icons item-icon">{item.icon}</span>
						{/if}
						<span class="item-name">{item.name}</span>
						{#if item.type === "page"}
							<span class="type-badge page">Navigation</span>
						{:else}
							<span class="type-badge game">Game</span>
						{/if}
					</div>
				{:else}
					<div class="no-results">
						<span class="material-icons">search_off</span>
						<p>No results found for "{searchQuery}"</p>
					</div>
				{/each}
			</div>

			<div class="palette-footer">
				<span class="tip"><span class="kbd">↑↓</span> Navigate</span>
				<span class="tip"><span class="kbd">Enter</span> Select</span>
				<span class="tip"><span class="kbd">Esc</span> Close</span>
			</div>
		</div>
	</button>
{/if}

<style lang="scss">
	.palette-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 15vh;
		z-index: 9999;
		border: none;
		outline: none;
		cursor: default;
	}

	.palette-container {
		width: 90%;
		max-width: 640px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-lg);
		box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6);
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.search-row {
		display: flex;
		align-items: center;
		padding: 16px 20px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);
		gap: 12px;

		.search-icon {
			color: var(--text-dim);
			font-size: 24px;
		}

		input {
			flex: 1;
			background: transparent;
			border: none;
			outline: none;
			color: var(--text-main);
			font-size: 1.1rem;
			font-weight: 600;

			&::placeholder {
				color: var(--text-dim);
			}
		}

		.esc-badge {
			font-size: 0.75rem;
			font-weight: 800;
			color: var(--text-dim);
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			padding: 4px 8px;
			border-radius: var(--radius-sm);
		}
	}

	.results-list {
		max-height: 360px;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 4px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 255, 255, 0.1);
			border-radius: var(--radius-pill);
		}
	}

	.result-item {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		border-radius: var(--radius-md);
		gap: 14px;
		cursor: pointer;
		outline: none;
		transition: background var(--transition-fast), transform var(--transition-fast);

		&.active {
			background: rgba(255, 255, 255, 0.06);
		}

		.item-icon {
			color: var(--text-dim);
			font-size: 20px;
		}

		.item-img-icon {
			width: 20px;
			height: 20px;
			object-fit: contain;
			border-radius: var(--radius-sm);
		}

		&.active .item-icon {
			color: var(--text-main);
		}

		.item-name {
			flex: 1;
			font-size: 0.95rem;
			font-weight: 700;
			color: var(--text-muted);
			text-align: left;
		}

		&.active .item-name {
			color: var(--text-main);
		}

		.type-badge {
			font-size: 0.7rem;
			font-weight: 800;
			text-transform: uppercase;
			padding: 2px 6px;
			border-radius: var(--radius-sm);
			letter-spacing: 0.5px;

			&.page {
				background: rgba(255, 255, 255, 0.05);
				color: var(--text-muted);
			}

			&.game {
				background: rgba(16, 185, 129, 0.1);
				color: var(--success);
			}
		}
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px 20px;
		color: var(--text-dim);
		gap: 8px;

		span {
			font-size: 32px;
		}

		p {
			margin: 0;
			font-size: 0.9rem;
			font-style: italic;
		}
	}

	.palette-footer {
		display: flex;
		align-items: center;
		padding: 12px 20px;
		background: var(--bg-surface);
		border-top: 2px solid rgba(255, 255, 255, 0.05);
		gap: 16px;

		.tip {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-dim);
			display: flex;
			align-items: center;
			gap: 4px;

			.kbd {
				background: rgba(255, 255, 255, 0.03);
				border: 1px solid rgba(255, 255, 255, 0.08);
				padding: 2px 5px;
				border-radius: var(--radius-sm);
				font-family: monospace;
			}
		}
	}
</style>
